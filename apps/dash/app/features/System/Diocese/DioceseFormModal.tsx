import { Form, Input, Modal } from "antd";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import type { Diocese, DioceseInput, MutationResult } from "./useDioceses";

interface DioceseFormModalProps {
  open: boolean;
  /** Kosong = mode tambah, terisi = mode ubah. */
  diocese: Diocese | null;
  onSubmit: (values: DioceseInput) => Promise<MutationResult>;
  onClose: () => void;
}

/** Modal tambah/ubah keuskupan. Error validasi dari API (422) ditampilkan di field-nya. */
export function DioceseFormModal({
  open,
  diocese,
  onSubmit,
  onClose,
}: DioceseFormModalProps) {
  const { t } = useTranslation();
  const [form] = Form.useForm<DioceseInput>();
  const [saving, setSaving] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  // Isi ulang tiap modal dibuka: form dipertahankan antar-buka karena `Modal` tidak unmount.
  useEffect(() => {
    if (!open) return;
    form.resetFields();
    form.setFieldsValue({ name: diocese?.name ?? "" });
    setFormError(null);
  }, [open, diocese, form]);

  async function handleOk() {
    const values = await form.validateFields().catch(() => null);
    if (!values) return;

    setSaving(true);
    setFormError(null);
    const result = await onSubmit({ name: values.name.trim() });
    setSaving(false);

    if (result.ok) return onClose();

    if ("name" in result.fields)
      form.setFields([{ name: "name", errors: [t("diocese.nameRequired")] }]);
    else
      setFormError(
        result.status === 403
          ? t("diocese.forbidden")
          : t("diocese.saveFailed"),
      );
  }

  return (
    <Modal
      open={open}
      title={diocese ? t("diocese.editTitle") : t("diocese.add")}
      okText={t("common.save")}
      cancelText={t("common.cancel")}
      confirmLoading={saving}
      onOk={handleOk}
      onCancel={onClose}
      destroyOnHidden
      maskClosable={!saving}
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={handleOk}
        requiredMark={false}
      >
        <Form.Item
          name="name"
          label={t("diocese.nameLabel")}
          rules={[
            {
              required: true,
              whitespace: true,
              message: t("diocese.nameRequired"),
            },
          ]}
          extra={
            formError ? (
              <span style={{ color: "var(--ant-color-error)" }}>
                {formError}
              </span>
            ) : undefined
          }
        >
          <Input autoFocus maxLength={200} />
        </Form.Item>
      </Form>
    </Modal>
  );
}
