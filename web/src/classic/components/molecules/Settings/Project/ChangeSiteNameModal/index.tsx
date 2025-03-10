import React from "react";

import Button from "@reearth/classic/components/atoms/Button";
import Divider from "@reearth/classic/components/atoms/Divider";
import Modal from "@reearth/classic/components/atoms/Modal";
import Text from "@reearth/classic/components/atoms/Text";
import TextBox from "@reearth/classic/components/atoms/TextBox";
import { useT } from "@reearth/services/i18n";
import { styled } from "@reearth/services/theme";
import { metricsSizes } from "@reearth/services/theme/metrics";

export type Props = {
  show?: boolean;
  onClose?: () => void;
  url?: string[];
  alias?: string;
  onAliasChange?: (value?: string | undefined) => void;
  handlePublish?: () => Promise<void>;
  disabled?: boolean;
};

const ChangeSiteNameModal: React.FC<Props> = ({
  show,
  onClose,
  url,
  alias,
  onAliasChange,
  handlePublish,
  disabled,
}) => {
  const t = useT();
  return (
    <Modal
      title={t("Change site name")}
      isVisible={show}
      onClose={onClose}
      size="sm"
      button1={
        <Button
          large
          text={t("Save")}
          buttonType="primary"
          onClick={handlePublish}
          disabled={disabled}
        />
      }
      button2={<Button large text={t("Cancel")} buttonType="secondary" onClick={onClose} />}>
      <Divider margin="5px" />
      <Content>
        <Text size="m">
          {t(
            "You are about to change the site name for your project. Only alphanumeric characters and hyphens are allows.",
          )}
        </Text>
        <TextBox prefix={url?.[0]} suffix={url?.[1]} value={alias ?? ""} onChange={onAliasChange} />
      </Content>
    </Modal>
  );
};

const Content = styled.div`
  * {
    margin: ${metricsSizes["m"]}px 0;
  }
`;

export default ChangeSiteNameModal;
