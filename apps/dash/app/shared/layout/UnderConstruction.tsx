import { ToolOutlined } from "@ant-design/icons";
import { Button, Result } from "antd";
import type { ReactNode } from "react";
import { Link } from "react-router";

interface UnderConstructionProps {
  /** Nama fitur/halaman, ditampilkan di judul. */
  title?: ReactNode;
  /** Tujuan tombol kembali. Kosongkan untuk menyembunyikan tombol. */
  backTo?: string;
}

/** Placeholder untuk halaman yang belum dibuat. */
export function UnderConstruction({
  title = "Sedang dibangun",
  backTo,
}: UnderConstructionProps) {
  return (
    <div className="glass grid place-items-center p-6">
      <Result
        icon={<ToolOutlined style={{ color: "var(--domus-primary)" }} />}
        title={title}
        subTitle="Fitur ini masih dalam tahap pengembangan."
        extra={
          backTo ? (
            <Link to={backTo}>
              <Button>Kembali</Button>
            </Link>
          ) : undefined
        }
      />
    </div>
  );
}
