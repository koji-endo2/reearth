package mongodoc

import (
	"github.com/reearth/reearth/server/pkg/workspace"
	"github.com/reearth/reearthx/mongox"
)

type PolicyDocument struct {
	ID                    workspace.PolicyID
	ProjectCount          *int
	PublishedProjectCount *int
	MemberCount           *int
	LayerCount            *int
	AssetStorageSize      *int64
}

func (d PolicyDocument) Model() *workspace.Policy {
	return workspace.NewPolicy(workspace.PolicyOption{
		ID:                    d.ID,
		ProjectCount:          d.ProjectCount,
		PublishedProjectCount: d.PublishedProjectCount,
		MemberCount:           d.MemberCount,
		LayerCount:            d.LayerCount,
		AssetStorageSize:      d.AssetStorageSize,
	})
}

type PolicyConsumer = mongox.SliceFuncConsumer[*PolicyDocument, *workspace.Policy]

func NewPolicyConsumer() *PolicyConsumer {
	return mongox.NewSliceFuncConsumer(func(d *PolicyDocument) (*workspace.Policy, error) {
		return d.Model(), nil
	})
}
