Feature: AI Assistant workspace, multi-session sidebar, Memory Drawer, and SSE chat streaming

  @org-member
  Scenario: Parish user sees Gabriel AI Assistant workspace on home page
    When I open "/"
    Then I see the AI Assistant workspace

  @org-member
  Scenario: Multi-session sidebar supports create, switch, rename, search, and delete
    Given I have two agent sessions "Jadwal Misa E2E" and "Info Lingkungan E2E"
    When I open "/"
    Then I see agent session "Jadwal Misa E2E" in sidebar
    And I see agent session "Info Lingkungan E2E" in sidebar
    When I rename agent session "Jadwal Misa E2E" to "Jadwal Misa Baru"
    Then I see agent session "Jadwal Misa Baru" in sidebar
    When I search agent sessions for "Lingkungan"
    Then I see agent session "Info Lingkungan E2E" in sidebar
    And I do not see agent session "Jadwal Misa Baru" in sidebar
    When I search agent sessions for ""
    And I delete agent session "Jadwal Misa Baru"
    Then I do not see agent session "Jadwal Misa Baru" in sidebar

  @org-member
  Scenario: Memory Drawer enforces personal CRUD and read-only parish FAQ for regular members
    Given the parish has a shared parish memory "Sekretariat Paroki buka Senin-Sabtu pukul 08.00-16.00 WIB"
    When I open "/"
    And I open the AI Memory Drawer
    And I create a personal memory in category "preference" with content "Panggil saya Kak Budi" and pinned true
    Then I see memory "Panggil saya Kak Budi"
    When I edit memory "Panggil saya Kak Budi" to "Panggil saya Mas Budi"
    Then I see memory "Panggil saya Mas Budi"
    When I switch memory scope tab to "FAQ & Pengetahuan Paroki"
    Then I see memory "Sekretariat Paroki buka Senin-Sabtu pukul 08.00-16.00 WIB"
    And I see the parish memory read-only notice
    When I switch memory scope tab to "Memori Pribadi"
    And I delete memory "Panggil saya Mas Budi"
    Then I do not see memory "Panggil saya Mas Budi"

  @org-member
  Scenario: SSE chat streams thinking indicator, tool execution badge, and assistant answer
    Given AI agent OpenAI provider is configured against the mock LLM server
    When I open "/"
    And I send AI chat prompt "Apa jadwal kegiatan paroki?"
    Then I see user chat message "Apa jadwal kegiatan paroki?"
    And I see thinking indicator for "Gabriel"
    And I see expandable tool execution badge for "event_list"
    And I see assistant reply containing "informasi kegiatan dan organisasi sudah tersedia"
