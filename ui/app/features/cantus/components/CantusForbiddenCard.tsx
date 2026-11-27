import { ArrowLeft, ShieldAlert } from "lucide-react";
import { Link } from "react-router";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";

export function CantusForbiddenCard() {
  return (
    <div className="mx-auto flex w-full max-w-2xl px-4 py-12 sm:px-6">
      <Card className="flex w-full flex-col items-center gap-3 p-6 text-center sm:p-8">
        <span className="flex size-12 items-center justify-center rounded-xl bg-danger-bg text-danger-fg">
          <ShieldAlert aria-hidden className="size-6" />
        </span>
        <h1 className="text-xl font-bold text-fg">Akses Ditolak</h1>
        <p className="text-sm text-muted">
          Hanya pengguna dengan peran Admin Nyanyian (cantus:admin) yang dapat
          menambah atau mengelola katalog nyanyian global.
        </p>
        <Link to="/cantus" className="mt-2 w-full no-underline sm:w-auto">
          <Button variant="outline" className="w-full sm:w-auto">
            <ArrowLeft aria-hidden className="size-4" />
            Kembali ke Katalog Nyanyian
          </Button>
        </Link>
      </Card>
    </div>
  );
}
