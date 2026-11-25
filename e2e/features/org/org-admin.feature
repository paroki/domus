Feature: Organization Admin menu visibility, /users roster CRUD, profile, and invites

  @org-admin
  Scenario: Org Admin sees restricted launcher, settings, /users directory, and /org create button hidden
    When I open "/"
    Then I see launcher tiles "Organisasi, Kegiatan, Pengguna"
    And I do not see launcher tiles "Referensi, Pengaturan"
    When I open "/ref"
    Then I see empty settings message
    When I open "/ref/paroki"
    Then I see the super admin restriction banner on paroki page
    When I open "/ref/profile"
    Then I see the paroki admin restriction banner
    When I open "/users"
    Then I see the users directory page
    And I do not see users directory controls "Pengguna Global, Admin & Anggota Paroki, Semua status, Semua Peran"
    When I open "/org"
    Then I see the orgs directory page
    And I do not see the create organization button

  @org-admin
  Scenario: Org Admin manages org member roster in /users/:id with last admin guard
    Given an org member with email "anggota.liturgi@domus.test" name "Anggota Liturgi E2E" position "Sekretaris E2E" exists in the active org and period
    When I open "/users"
    Then I see the users directory page
    When I open user detail "Admin Organisasi E2E"
    And I toggle active org membership role via "Jadikan Anggota"
    Then I see users error "Minimal harus ada satu Admin Organisasi aktif yang tersisa."
    When I toggle active org membership status via "Nonaktifkan"
    Then I see users error "Minimal harus ada satu Admin Organisasi aktif yang tersisa."
    When I go back to the users directory
    And I open user detail "Anggota Liturgi E2E"
    Then I see active org membership with role "Anggota"
    And I see active org membership metadata "Sekretaris E2E"
    And I see active org membership metadata "Periode 2025-2028"
    When I edit active org membership position to "Bendahara"
    Then I see active org membership metadata "Bendahara"
    When I toggle active org membership role via "Jadikan Admin"
    Then I see active org membership with role "Admin Organisasi"
    When I toggle active org membership role via "Jadikan Anggota"
    Then I see active org membership with role "Anggota"
    When I toggle active org membership status via "Nonaktifkan"
    Then I see active org membership status "Nonaktif"
    When I toggle active org membership status via "Aktifkan"
    Then I see active org membership status "Aktif"
    When I remove active org membership
    And I go back to the users directory
    Then I do not see user "anggota.liturgi@domus.test"

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
