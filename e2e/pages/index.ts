import type { Page } from "playwright";
import { AgentWorkspacePage } from "./agent/agent-workspace.page";
import { BasePage } from "./base.page";
import { CantusPage } from "./cantus/cantus.page";
import { EventDetailPage } from "./event/event-detail.page";
import { EventsPage } from "./event/events.page";
import { OrgDetailPage } from "./org/org-detail.page";
import { OrgJoinPage } from "./org/org-join.page";
import { OrgsPage } from "./org/orgs.page";
import { KevikepanSipilPage } from "./ref/kevikepan-sipil.page";
import { LookupsPage } from "./ref/lookups.page";
import { ParokiPage } from "./ref/paroki.page";
import { ParokiProfilePage } from "./ref/paroki-profile.page";
import { SettingsPage } from "./ref/settings.page";
import { WilayahPage } from "./ref/wilayah.page";
import { AIAgentSettingsPage } from "./settings/ai-agent.page";
import { AIProviderSettingsPage } from "./settings/ai-provider.page";
import { SystemSettingsPage } from "./settings/system-settings.page";
import { HomePage } from "./shell/home.page";
import { UmatPage } from "./umat/umat.page";
import { LoginPage } from "./user/login.page";
import { MePage } from "./user/me.page";
import { OnboardingPage } from "./user/onboarding.page";
import { UsersPage } from "./user/users.page";

export function createPages(page: Page) {
  return {
    base: new BasePage(page),
    event: {
      events: new EventsPage(page),
      detail: new EventDetailPage(page),
    },
    org: {
      orgs: new OrgsPage(page),
      detail: new OrgDetailPage(page),
      join: new OrgJoinPage(page),
    },
    ref: {
      settings: new SettingsPage(page),
      paroki: new ParokiPage(page),
      kevikepanSipil: new KevikepanSipilPage(page),
      lookups: new LookupsPage(page),
      profile: new ParokiProfilePage(page),
      wilayah: new WilayahPage(page),
    },
    shell: {
      home: new HomePage(page),
    },
    umat: new UmatPage(page),
    cantus: new CantusPage(page),
    settings: new SystemSettingsPage(page),
    aiProvider: new AIProviderSettingsPage(page),
    aiAgent: new AIAgentSettingsPage(page),
    agent: {
      workspace: new AgentWorkspacePage(page),
    },
    user: {
      login: new LoginPage(page),
      users: new UsersPage(page),
      onboarding: new OnboardingPage(page),
      me: new MePage(page),
    },
  };
}

export type AppPages = ReturnType<typeof createPages>;
