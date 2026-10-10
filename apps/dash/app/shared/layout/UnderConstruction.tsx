import { ToolOutlined } from "@ant-design/icons";
import { Button, Result } from "antd";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";

interface UnderConstructionProps {
  /** Nama fitur/halaman, ditampilkan di judul. Bawaan: teks terjemahan. */
  title?: ReactNode;
  /** Tujuan tombol kembali. Kosongkan untuk menyembunyikan tombol. */
  backTo?: string;
}

/** Placeholder untuk halaman yang belum dibuat. */
export function UnderConstruction({ title, backTo }: UnderConstructionProps) {
  const { t } = useTranslation();

  return (
    <div className="glass grid place-items-center p-6">
      <Result
        icon={<ToolOutlined style={{ color: "var(--domus-primary)" }} />}
        title={title ?? t("underConstruction.title")}
        subTitle={t("underConstruction.description")}
        extra={
          backTo ? (
            <Link to={backTo}>
              <Button>{t("common.back")}</Button>
            </Link>
          ) : undefined
        }
      />
    </div>
  );
}
