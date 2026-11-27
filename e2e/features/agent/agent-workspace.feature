Feature: Parish user AI Assistant workspace

  @org-member
  Scenario: Parish user chats with Gabriel, sees Mikael handoff, aborts slow stream, and manages sessions and memories
    Given AI provider and agent models are configured with mock LLM
    When I open "/"
    Then I see launcher tiles "Asisten AI"
    When I open launcher tile "Asisten AI"
    Then I am on "/agent"
    And I see the AI Assistant workspace
    When I send AI chat message "Tolong cek jadwal kegiatan paroki"
    Then I see AI agent handoff from "Gabriel" to "Mikael"
    And I see AI tool badge "event_list"
    And I see AI assistant reply containing "Berdasarkan data paroki"
    And the AI session URL persists across page reload with reply "Berdasarkan data paroki"
    When I send slow AI chat message "[SLOW_STREAM] cek jadwal lagi" and click Hentikan
    And I send AI chat message "Halo lagi"
    Then I see AI assistant reply containing "Halo!"
    When I rename the active AI session to "Jadwal Misa Paroki"
    And I delete the AI session "Jadwal Misa Paroki"
    Then I am on "/agent"
    When I open the AI memory drawer
    And I add an AI memory "Saya lebih suka misa pagi pukul 06:00"
    And I filter AI memories by category "Fakta"
    And I search AI memories for "misa pagi" expecting "Saya lebih suka misa pagi pukul 06:00"
    And I filter AI memories by category "Semua"
    And I edit AI memory "Saya lebih suka misa pagi pukul 06:00" to "Saya suka misa pagi pukul 06:00 WIB"
    And I delete AI memory "Saya suka misa pagi pukul 06:00 WIB"

  @org-member
  Scenario: Parish user delegates to Yosef and Sesilia and uses quick-reply options
    Given AI provider and agent models are configured with mock LLM
    When I open "/agent"
    Then I see the AI Assistant workspace
    When I send AI chat message "Tolong cek data umat dan kartu keluarga"
    Then I see AI agent handoff from "Gabriel" to "Yosef"
    And I see AI tool badge "umat_list_keluarga"
    And I see AI assistant reply containing "Berdasarkan data paroki"
    When I send AI chat message "Carikan lagu pujian cantus"
    Then I see AI agent handoff from "Gabriel" to "Sesilia"
    And I see AI tool badge "cantus_search"
    And I see AI assistant reply containing "Berdasarkan data paroki"
    When I send AI chat message "[ASK_USER] tanya konfirmasi jadwal misa"
    Then I see AI tool badge "agent_ask_user"
    When I click AI quick-reply option "Misa Minggu Pagi 08:00"
    Then I see AI assistant reply containing "Halo!"

  @org-member
  Scenario: Parish user uses global bottom dock with page context, auto-navigation, and Buka Penuh
    Given AI provider and agent models are configured with mock LLM
    When I open "/event"
    And I open the AI bottom dock
    And I send AI bottom dock message "[ECHO_CONTEXT] ringkas halaman ini"
    Then I see AI bottom dock reply containing "- Active Page: /event"
    When I send AI bottom dock message "[NAVIGATE_AUTO] buka halaman organisasi"
    Then I am on "/org"
    And the AI bottom dock remains open
    And I see AI bottom dock reply containing "- Active Page: /event"
    When I click Buka Penuh in the AI bottom dock
    Then I see AI assistant reply containing "- Active Page: /event"
