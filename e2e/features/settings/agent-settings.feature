Feature: Super Admin system settings for communication channels, LLM providers, and AI agent personas

  @parish-admin
  Scenario: Non-Super Admin sees restriction banner on /settings
    When I open "/settings"
    Then I see the system settings forbidden banner

  @super-admin
  Scenario: Super Admin separates channel settings, syncs OpenAI provider models, and configures agent personas
    When I open "/settings"
    Then I see the system settings page
    And I see channel groups "Telegram Bot, GOWA (WhatsApp)" without raw agent keys
    When I switch settings tab to "Penyedia LLM"
    And I configure provider "openai" with key "e2e-test-api-key" and mock LLM base URL
    And I sync models for provider "openai" labeled "OpenAI"
    Then I see provider "OpenAI" model catalog containing "gpt-4.1-mini, gpt-4.1, text-embedding-3-small"
    When I switch settings tab to "Agen & Embedding"
    And I configure agent "Gabriel" with provider "openai", model "gpt-4.1-mini", effort "medium", and instruction "Instruksi Gabriel E2E"
    And I configure agent "Mikael" with provider "openai", model "gpt-4.1-mini", and effort "low"
    And I configure agent "Rafael" with provider "openai", model "gpt-4.1-mini", and effort "low"
    And I configure embedding with provider "openai" and model "text-embedding-3-small"
    And I save agent configuration
    When I open "/settings?tab=agents"
    Then I see agent "Gabriel" configured with provider "openai", model "gpt-4.1-mini", effort "medium", and instruction "Instruksi Gabriel E2E"
    When I reset instruction for agent "Gabriel"
    Then I see agent "Gabriel" instruction restored from default
