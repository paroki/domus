Feature: Super Admin system settings for communication channels and LLM providers

  @paroki-admin
  Scenario: Non-Super Admin sees restriction banner on /settings/system
    When I open "/settings/system"
    Then I see the system settings forbidden banner

  @super-admin
  Scenario: Super Admin navigates launcher routes, configures channels, and syncs OpenAI models
    When I open "/settings/system"
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
    When I open settings launcher "Kembali ke Pengaturan"
    And I open settings launcher "AI Agent"
    And I bind AI agent models gabriel "gpt-4.1-mini", mikael "gpt-4.1", rafael "gpt-4.1", embedding "text-embedding-3-small"
    And I customize Gabriel prompt to "Instruksi kustom E2E dengan {curly_braces}."
    And I reset Gabriel prompt to default
    When I open settings launcher "Kembali ke Pengaturan"
    And I open settings launcher "AI Provider"
    Then deleting provider "OpenAI" is blocked because it is in use
