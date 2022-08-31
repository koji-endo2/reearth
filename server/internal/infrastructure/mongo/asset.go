package mongo

import (
	"context"
	"fmt"
	"regexp"

	"github.com/reearth/reearth/server/internal/infrastructure/mongo/mongodoc"
	"github.com/reearth/reearth/server/internal/usecase/repo"
	"github.com/reearth/reearth/server/pkg/asset"
	"github.com/reearth/reearth/server/pkg/id"
	"github.com/reearth/reearthx/log"
	"github.com/reearth/reearthx/mongox"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Asset struct {
	client *mongox.ClientCollection
	f      repo.WorkspaceFilter
}

func NewAsset(client *mongox.Client) *Asset {
	r := &Asset{client: client.WithCollection("asset")}
	r.init()
	return r
}

func (r *Asset) Filtered(f repo.WorkspaceFilter) repo.Asset {
	return &Asset{
		client: r.client,
		f:      r.f.Merge(f),
	}
}

func (r *Asset) FindByID(ctx context.Context, id id.AssetID) (*asset.Asset, error) {
	return r.findOne(ctx, bson.M{
		"id": id.String(),
	})
}

func (r *Asset) FindByIDs(ctx context.Context, ids id.AssetIDList) ([]*asset.Asset, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	res, err := r.find(ctx, bson.M{
		"id": bson.M{"$in": ids.Strings()},
	})
	if err != nil {
		return nil, err
	}
	return filterAssets(ids, res), nil
}

func (r *Asset) FindByWorkspace(ctx context.Context, id id.WorkspaceID, uFilter repo.AssetFilter) ([]*asset.Asset, *usecasex.PageInfo, error) {
	if !r.f.CanRead(id) {
		return nil, usecasex.EmptyPageInfo(), nil
	}

	var filter any = bson.M{
		"team": id.String(),
	}

	if uFilter.Keyword != nil {
		filter = mongox.And(filter, "name", bson.M{
			"$regex": primitive.Regex{Pattern: fmt.Sprintf(".*%s.*", regexp.QuoteMeta(*uFilter.Keyword)), Options: "i"},
		})
	}

	return r.paginate(ctx, filter, uFilter.Sort, uFilter.Pagination)
}

func (r *Asset) TotalSizeByWorkspace(ctx context.Context, wid id.WorkspaceID) (int64, error) {
	if !r.f.CanRead(wid) {
		return 0, repo.ErrOperationDenied
	}

	c, err := r.client.Client().Aggregate(ctx, []bson.M{
		{"$match": bson.M{"team": wid.String()}},
		{"$group": bson.M{"_id": nil, "size": bson.M{"$sum": "$size"}}},
	})
	if err != nil {
		return 0, rerror.ErrInternalBy(err)
	}
	defer func() {
		_ = c.Close(ctx)
	}()
	_ = c.Next(ctx)
	type resp struct {
		Size int64
	}
	var res resp
	if err := c.Decode(&res); err != nil {
		return 0, rerror.ErrInternalBy(err)
	}
	return res.Size, nil
}

func (r *Asset) Save(ctx context.Context, asset *asset.Asset) error {
	if !r.f.CanWrite(asset.Workspace()) {
		return repo.ErrOperationDenied
	}
	doc, id := mongodoc.NewAsset(asset)
	return r.client.SaveOne(ctx, id, doc)
}

func (r *Asset) Remove(ctx context.Context, id id.AssetID) error {
	return r.client.RemoveOne(ctx, r.writeFilter(bson.M{
		"id": id.String(),
	}))
}

func (r *Asset) init() {
	i := r.client.CreateIndex(context.Background(), []string{"team"}, []string{"id"})
	if len(i) > 0 {
		log.Infof("mongo: %s: index created: %s", "asset", i)
	}
}

func (r *Asset) paginate(ctx context.Context, filter any, sort *asset.SortType, pagination *usecasex.Pagination) ([]*asset.Asset, *usecasex.PageInfo, error) {
	var sortstr *string
	if sort != nil {
		sortstr2 := string(*sort)
		sortstr = &sortstr2
	}

	c := mongodoc.NewAssetConsumer()
	pageInfo, err := r.client.Paginate(ctx, r.readFilter(filter), sortstr, pagination, c)
	if err != nil {
		return nil, nil, rerror.ErrInternalBy(err)
	}
	return c.Result, pageInfo, nil
}

func (r *Asset) find(ctx context.Context, filter any) ([]*asset.Asset, error) {
	c := mongodoc.NewAssetConsumer()
	if err2 := r.client.Find(ctx, r.readFilter(filter), c); err2 != nil {
		return nil, rerror.ErrInternalBy(err2)
	}
	return c.Result, nil
}

func (r *Asset) findOne(ctx context.Context, filter any) (*asset.Asset, error) {
	c := mongodoc.NewAssetConsumer()
	if err := r.client.FindOne(ctx, r.readFilter(filter), c); err != nil {
		return nil, err
	}
	return c.Result[0], nil
}

func filterAssets(ids []id.AssetID, rows []*asset.Asset) []*asset.Asset {
	res := make([]*asset.Asset, 0, len(ids))
	for _, id := range ids {
		var r2 *asset.Asset
		for _, r := range rows {
			if r.ID() == id {
				r2 = r
				break
			}
		}
		res = append(res, r2)
	}
	return res
}

func (r *Asset) readFilter(filter any) any {
	return applyWorkspaceFilter(filter, r.f.Readable)
}

func (r *Asset) writeFilter(filter any) any {
	return applyWorkspaceFilter(filter, r.f.Writable)
}
