Feature: Parish Liturgy Schedules, 6 Ministries, Officers, and Songs

  @paroki-admin
  Scenario: Parish Admin delegates liturgi:admin, configures 6 ministries, creates Sunday Mass schedule, and parishioner submits RSVP
    Given a paroki member "liturgi-admin.e2e@domus.test" named "Liturgi Admin E2E" exists in the active paroki
    And a paroki member "lektor-coord.e2e@domus.test" named "Koordinator Lektor E2E" exists in the active paroki
    And a paroki member "umat-rsvp.e2e@domus.test" named "Umat RSVP E2E" exists in the active paroki
    And a parish choir org "Kor Paroki E2E" coded "kor-paroki-e2e" exists in the active paroki
    When I open "/users"
    And I open user detail for "liturgi-admin.e2e@domus.test"
    And I switch to user detail tab "Paroki"
    And I grant liturgi admin on the active paroki
    Then I see the liturgi admin badge on the active paroki
    When I sign in as "liturgi-admin.e2e@domus.test" named "Liturgi Admin E2E" on the active paroki
    And I open "/liturgi/ministries"
    And I map liturgi ministry "Lektor" to leader "Koordinator Lektor E2E" with query "Koordinator Lektor"
    And I map liturgi ministry "Kor" to org "Kor Paroki E2E"
    And I open "/liturgi/new"
    And I create a liturgi schedule titled "Misa Minggu Pagi E2E" at "Gereja Utama" start "2099-11-08T08:00" end "2099-11-08T09:30"
    And I open "/liturgi"
    Then I see liturgi schedule "Misa Minggu Pagi E2E" in desktop 7-column matrix and mobile card stack
    When I open "/event"
    And I open event detail "Misa Minggu Pagi E2E"
    Then I see event liturgi shortcut button
    When I click event liturgi shortcut button
    Then I see liturgi detail "Misa Minggu Pagi E2E"
    When I sign in as "umat-rsvp.e2e@domus.test" named "Umat RSVP E2E" on the active paroki
    And I open the active liturgi schedule
    And I submit liturgi RSVP "Hadir"
    Then I see liturgi RSVP status "Hadir"

  @paroki-admin
  Scenario: Ministry leader assigns officers within role scope and approves officer substitute proposal
    Given a liturgi setup with lektor coordinator "lektor-coord2.e2e@domus.test" named "Koordinator Lektor Dua", lectors "Lektor Satu E2E" and "Lektor Dua E2E", and Mass "Misa Minggu Siang E2E"
    When I sign in as "lektor-coord2.e2e@domus.test" named "Koordinator Lektor Dua" on the active paroki
    And I open the active liturgi schedule
    And I open add liturgi petugas form
    Then I see only role "Lektor" in the liturgi role select
    When I assign liturgi petugas slot "Bacaan I" to member "Lektor Satu E2E" with query "Lektor Satu"
    And I open add liturgi petugas form
    And I assign liturgi petugas slot "Bacaan II" to member "Lektor Dua E2E" with query "Lektor Dua"
    Then I see "Lektor Satu E2E" as the main officer for liturgi slot "Bacaan I"
    And I see "Lektor Dua E2E" as the main officer for liturgi slot "Bacaan II"
    When I sign in as "lektor1.e2e@domus.test" named "Lektor Satu E2E" on the active paroki
    And I open "/liturgi"
    And I filter liturgi schedules by "Jadwal Tugas Saya"
    And I open liturgi schedule "Misa Minggu Siang E2E" from the matrix
    And I decline liturgi slot "Bacaan I" with note "Sedang di luar kota" and propose substitute "Lektor Dua E2E" with query "Lektor Dua"
    And I sign in as "lektor-coord2.e2e@domus.test" named "Koordinator Lektor Dua" on the active paroki
    And I open the active liturgi schedule
    Then I see substitute proposal "Lektor Dua E2E" on liturgi slot "Bacaan I"
    When I approve substitute on liturgi slot "Bacaan I"
    Then I see "Lektor Dua E2E" as the main officer for liturgi slot "Bacaan I"

  @paroki-admin
  Scenario: Choir curates liturgical songs from Cantus catalog and manual entry with 409 CANTUS_IN_USE guard
    Given a liturgi song setup with choir "Kor Gloria E2E" led by "kor-admin.e2e@domus.test", Mass "Misa Minggu Sore E2E", and cantus song "Tuhan Kasihanilah Kami E2E"
    When I sign in as "kor-admin.e2e@domus.test" named "Kor Admin E2E" on the active paroki
    And I open the active liturgi schedule
    And I open add liturgi lagu form
    And I add liturgi song from cantus "Tuhan Kasihanilah Kami E2E" for bagian "pembuka" expecting sumber "PS" nomor "PS-351" nadaDasar "F"
    And I open add liturgi lagu form
    And I add manual liturgi song titled "Lagu Penutup Manual E2E" for bagian "penutup"
    Then I see catalog-linked liturgi song "Tuhan Kasihanilah Kami E2E" and manual song "Lagu Penutup Manual E2E"
    And deleting the referenced cantus song is rejected with 409 CANTUS_IN_USE
