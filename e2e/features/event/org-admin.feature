Feature: Organization Admin public and private events lifecycle

  @org-admin
  Scenario: Org Admin creates public and private events, validates time ranges, and filters/searches on /event
    When I open "/event"
    Then I see the events page
    When I submit an event with title "Kegiatan Invalid" start "2099-06-15T10:00" end "2099-06-15T09:00" visibility "Publik Paroki" location "Aula"
    Then I see event form error "Waktu selesai harus setelah waktu mulai, dan batas RSVP tidak boleh melewati waktu selesai."
    When I create an event with title "Misa Syukur Paroki E2E" start "2099-06-15T09:00" end "2099-06-15T11:00" visibility "Publik Paroki" location "Gereja Utama"
    And I create an event with title "Rapat Internal Seksi Liturgi E2E" start "2099-06-16T18:00" end "2099-06-16T20:00" visibility "Internal Organisasi" location "Ruang Rapat DPP"
    Then I see event card "Misa Syukur Paroki E2E" with visibility "Publik Paroki"
    And I see event card "Rapat Internal Seksi Liturgi E2E" with visibility "Internal Organisasi"
    When I filter events by visibility "Internal Organisasi"
    Then I see event card "Rapat Internal Seksi Liturgi E2E" with visibility "Internal Organisasi"
    And I do not see event card "Misa Syukur Paroki E2E"
    When I filter events by visibility "Publik Paroki"
    Then I see event card "Misa Syukur Paroki E2E" with visibility "Publik Paroki"
    And I do not see event card "Rapat Internal Seksi Liturgi E2E"
    When I filter events by visibility "Semua Visibilitas"
    And I search events for "Ruang Rapat DPP"
    Then I see event card "Rapat Internal Seksi Liturgi E2E" with visibility "Internal Organisasi"

  @org-admin
  Scenario: Org Admin edits, cancels, reschedules, and deletes an event at /event/:eventId
    When I open "/event"
    Then I see the events page
    When I create an event with title "Latihan Koor Gabungan E2E" start "2099-07-01T17:00" end "2099-07-01T19:00" visibility "Internal Organisasi" location "Aula Lama"
    And I open event detail "Latihan Koor Gabungan E2E"
    Then I see event detail "Latihan Koor Gabungan E2E"
    When I edit the event to title "Latihan Koor Raya E2E" location "Aula Baru Paroki" visibility "Publik Paroki"
    Then I see event detail "Latihan Koor Raya E2E"
    And I see event detail location "Aula Baru Paroki"
    And I see event badges "Publik Paroki, Akan Datang"
    When I cancel the event
    Then I see event badges "Dibatalkan"
    And I see event RSVP closed message
    When I reschedule the event
    Then I see event badges "Akan Datang"
    And I see event RSVP open buttons
    When I delete the event
    Then I am on "/event"
    And I do not see event card "Latihan Koor Raya E2E"
