import { Typography } from "antd";
import { Link } from "react-router";
import { modulePath, modules } from "~/shared/modules/registry";
import type { Route } from "../../routes/+types/_app._index";

export function meta(_: Route.MetaArgs) {
  return [{ title: "Beranda | Domus" }];
}

/** App launcher: grid ikon semua modul, seperti home menu Odoo. */
export default function LauncherPage() {
  return (
    <section
      aria-labelledby="launcher-title"
      className="mx-auto w-full max-w-4xl px-4 py-10"
    >
      <Typography.Title
        id="launcher-title"
        level={2}
        style={{ margin: 0, fontWeight: 500 }}
      >
        Semua modul
      </Typography.Title>
      <Typography.Paragraph
        style={{ margin: "6px 0 0", color: "var(--domus-text-secondary)" }}
      >
        Pilih modul untuk mulai bekerja.
      </Typography.Paragraph>

      <ul className="mt-6 grid list-none grid-cols-2 gap-3 p-0 sm:grid-cols-3 md:grid-cols-5">
        {modules.map(({ id, name, icon: Icon }) => (
          <li key={id}>
            <Link
              to={modulePath(id)}
              className="glass launcher-tile flex h-full flex-col items-center gap-3 p-5 text-center"
            >
              <span aria-hidden="true" className="icon-tile">
                <Icon />
              </span>
              <span style={{ fontWeight: 500 }}>{name}</span>
            </Link>
          </li>
        ))}
      </ul>
    </section>
  );
}
