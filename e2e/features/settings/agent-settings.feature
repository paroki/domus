Feature: Super Admin system settings for communication channels, LLM providers, and AI agent personas

  @parish-admin
  Scenario: Non-Super Admin sees restriction banner on /settings
    When I open "/settings"
    Then I see the system settings forbidden banner

  @super-admin
  Scenario: Super Admin navigates launcher routes, configures channels, syncs OpenAI models, and configures agent personas
    When I open "/settings"
    Then I see the system settings page
    When I open settings launcher "Telegram"
    Then I see channel groups "Telegram Bot" without raw agent keys
    When I open settings launcher "Kembali ke Pengaturan"
    And I open settings launcher "WhatsApp"
    Then I see channel groups "GOWA (WhatsApp)" without raw agent keys
    When I open settings launcher "Kembali ke Pengaturan"
    And I open settings launcher "AI Provider"
    And I configure provider "openai" with key "e2e-test-api-key" and mock LLM base URL
    And I sync models for provider "openai" labeled "OpenAI"
    Then I see provider "OpenAI" model catalog containing "gpt-4.1-mini, gpt-4.1, text-embedding-3-small"
    When I open "/settings/ai-agent"
    And I configure embedding with provider "openai" and model "text-embedding-3-small"
    And I save embedding configuration
    When I open settings launcher "Gabriel"
    And I configure agent "Gabriel" with provider "openai", model "gpt-4.1-mini", effort "medium", and instruction "Instruksi Gabriel E2E"
    And I save agent configuration
    When I open "/settings/ai-agent/gabriel"
    Then I see agent "Gabriel" configured with provider "openai", model "gpt-4.1-mini", effort "medium", and instruction "Instruksi Gabriel E2E"
    When I reset instruction for agent "Gabriel"
    Then I see agent "Gabriel" instruction restored from default
