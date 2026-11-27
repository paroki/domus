import { ArrowLeft } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Link, useNavigate } from "react-router";
import i18n from "~/shared/i18n";
import { buttonClass } from "~/shared/ui/Button";
import { InductiveFormSection, InductivePageHeader } from "~/shared/ui/iui";
import { useMe } from "../api/queries";
import { WhatsAppLinkForm } from "../components/WhatsAppLinkForm";
import { ME_ROUTES } from "../model";
import type { Route } from "./+types/WhatsAppLinkPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("me.whatsappPageTitle")} · Domus` }];
}

export default function WhatsAppLinkPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data: me } = useMe();

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: t("me.homeBreadcrumb"), to: "/" },
          { label: t("me.title"), to: ME_ROUTES.index },
          { label: t("me.whatsappPageTitle") },
        ]}
        title={t("me.whatsappPageTitle")}
        subtitle={t("me.whatsappPageSubtitle")}
      />

      <InductiveFormSection
        title={t("me.whatsappPageTitle")}
        description={t("me.whatsappPageSubtitle")}
      >
        <WhatsAppLinkForm
          initialPhone={me?.phone}
          onLinked={() => {
            navigate(ME_ROUTES.index);
          }}
        />
      </InductiveFormSection>

      <div>
        <Link
          to={ME_ROUTES.index}
          className={`${buttonClass("outline")} w-full sm:w-auto`}
        >
          <ArrowLeft aria-hidden className="size-4" />
          {t("me.backToProfile")}
        </Link>
      </div>
    </div>
  );
}
