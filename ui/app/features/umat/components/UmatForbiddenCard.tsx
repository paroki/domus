import { ShieldAlert } from "lucide-react";
import { Link } from "react-router";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";

export function UmatForbiddenCard({
  title = "Akses Data Umat Diperlukan",
  description = "Halaman ini hanya dapat diakses oleh Admin Data Umat pada paroki yang aktif.",
}: {
  title?: string;
  description?: string;
}) {
  return (
    <div className="mx-auto flex w-full max-w-2xl flex-col items-center px-4 py-16">
      <Card className="flex w-full flex-col items-center gap-3 p-8 text-center">
        <span className="flex size-12 items-center justify-center rounded-xl bg-danger-bg text-danger-fg">
          <ShieldAlert aria-hidden className="size-6" />
        </span>
        <h1 className="text-xl font-bold text-fg">{title}</h1>
        <p className="max-w-md text-sm text-muted">{description}</p>
        <Link to="/" className="mt-2 no-underline">
          <Button variant="outline">Kembali ke beranda</Button>
        </Link>
      </Card>
    </div>
  );
}
