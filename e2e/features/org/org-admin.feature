Feature: Organization Admin menu visibility, /users roster CRUD, profile, and invites

  @org-admin
  Scenario: Org Admin sees restricted launcher, settings, /users scope, and /org create button hidden
    When I open "/"
    Then I see launcher tiles "Organisasi, Kegiatan, Pengguna"
    And I do not see launcher tiles "Pengaturan Paroki"
    When I open "/settings"
    Then I see empty settings message
    When I open "/settings/parishes"
    Then I see the super admin restriction banner on parishes page
    When I open "/settings/profile"
    Then I see the parish admin restriction banner
    When I open "/users"
    Then I see users directory scopes "Pengurus & Anggota Organisasi"
    And I do not see users directory scopes "Pengguna Global, Admin & Anggota Paroki"
    When I open "/org"
    Then I see the orgs directory page
    And I do not see the create organization button

  @org-admin
  Scenario: Org Admin manages org member roster in /users with last admin guard
    When I open "/users"
    Then I see the users directory page
    When I toggle org member "Admin Organisasi E2E" role via "Jadikan Anggota"
    Then I see users error "Minimal harus ada satu Admin Organisasi aktif yang tersisa."
    When I toggle org member "Admin Organisasi E2E" status via "Nonaktifkan"
    Then I see users error "Minimal harus ada satu Admin Organisasi aktif yang tersisa."
    When I add org member with email "anggota.liturgi@domus.test" name "Anggota Liturgi E2E" phone "+6281311112222" role "Anggota" position "Sekretaris E2E" period "Periode 2025-2028"
    Then I see org member "Anggota Liturgi E2E" with role "Anggota"
    And I see org member "Anggota Liturgi E2E" metadata "Sekretaris E2E"
    And I see org member "Anggota Liturgi E2E" metadata "Periode 2025-2028"
    When I edit org member "Anggota Liturgi E2E" position to "Bendahara"
    Then I see org member "Anggota Liturgi E2E" metadata "Bendahara"
    When I toggle org member "Anggota Liturgi E2E" role via "Jadikan Admin"
    Then I see org member "Anggota Liturgi E2E" with role "Admin Organisasi"
    When I toggle org member "Anggota Liturgi E2E" role via "Jadikan Anggota"
    Then I see org member "Anggota Liturgi E2E" with role "Anggota"
    When I toggle org member "Anggota Liturgi E2E" status via "Nonaktifkan"
    Then I see org member "Anggota Liturgi E2E" status "Nonaktif"
    When I toggle org member "Anggota Liturgi E2E" status via "Aktifkan"
    Then I see org member "Anggota Liturgi E2E" status "Aktif"
    When I remove org member "Anggota Liturgi E2E"
    Then I do not see org member "Anggota Liturgi E2E"

  @org-admin
  Scenario: Org Admin updates organization profile, views read-only roster, and manages invite links
    When I open "/org"
    And I open organization detail "Seksi Liturgi E2E"
    Then I see organization detail "Seksi Liturgi E2E"
    And I see org edit button and no delete button
    And I see read-only org roster with member "Admin Organisasi E2E"
    When I update organization description to "Tim pelayanan liturgi dan tata perayaan ekaristi"
    Then I see organization description "Tim pelayanan liturgi dan tata perayaan ekaristi"
    When I create org invite with role "Anggota (member)" position "Anggota Paduan Suara" max uses 5
    Then I see org invite with status "Aktif" position "Anggota Paduan Suara" usage "Terpakai: 0 / 5"
    When I revoke the org invite
    Then I see org invite with status "Dicabut"
