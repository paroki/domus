Feature: Organization Admin menu visibility, /users roster CRUD, profile, and invites

  @org-admin
  Scenario: Org Admin sees restricted launcher, settings, /users directory, and /org create button hidden
    When I open "/"
    Then I see launcher tiles "Organisasi, Kegiatan, Pengguna"
    And I do not see launcher tiles "Pengaturan"
    When I open "/settings"
    Then I see empty settings message
    When I open "/settings/gerejawi/paroki"
    Then I see the super admin restriction banner on paroki page
    When I open "/settings/paroki/profile"
    Then I see the paroki admin restriction banner
    When I open "/users"
    Then I see the users directory page
    And I do not see users directory controls "Pengguna Global, Admin & Anggota Paroki, Semua status, Semua Peran"
    When I open "/org"
    Then I see the orgs directory page
    And I do not see the create organization button

  @org-admin
  Scenario: Org Admin manages org member roster in /users/:id with last admin guard
    Given an org member with email "anggota.liturgi@domus.test" name "Anggota Liturgi E2E" exists in the active org
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
    When I create org invite with role "Anggota (member)" max uses 5
    Then I see org invite with status "Aktif" usage "Terpakai: 0 / 5"
    When I revoke the org invite
    Then I see org invite with status "Dicabut"

  @org-admin
  Scenario: Org Admin manages org_periode lifecycle with overlap guard, active rotation, and open/closed toggle
    When I open "/org"
    And I open organization detail "Seksi Liturgi E2E"
    Then I see org period "Periode 2025-2028" with status "Terbuka" and current badge
    When I open the new org period form and submit "Periode Overlap" from "2026-01-01" to "2027-12-31"
    Then I see org error "Rentang tanggal periode bertumpuk dengan periode lain yang sudah ada."
    When I resubmit the org period form with name "Periode 2029-2031" from "2029-01-01" to "2031-12-31"
    Then I see org period "Periode 2029-2031" with status "Terbuka"
    When I make org period "Periode 2029-2031" active
    Then I see org period "Periode 2029-2031" with status "Terbuka" and current badge
    And I see org period "Periode 2025-2028" without current badge
    When I toggle org period "Periode 2025-2028" via "Tutup"
    Then I see org period "Periode 2025-2028" with status "Ditutup"
    When I delete org period "Periode 2029-2031"
    Then I do not see org period "Periode 2029-2031"

  @org-admin
  Scenario: Org Admin manages hierarchical org_jabatan tree with umat lookup, rangkap jabatan, and ORG_IN_USE guards
    Given a parishioner named "Yohanes Budi" exists in the active parish
    When I open "/org"
    And I open organization detail "Seksi Liturgi E2E"
    And I create root org jabatan "Ketua" assigned to org member "Admin Organisasi E2E"
    Then I see org jabatan "Ketua" held by "Admin Organisasi E2E"
    When I create child org jabatan "Bendahara" under "Ketua" assigned to umat "Yohanes Budi"
    Then I see child org jabatan "Bendahara" held by "Yohanes Budi"
    When I create child org jabatan "Dirigen" under "Ketua" with manual name "Ibu Maria"
    Then I see child org jabatan "Dirigen" held by "Ibu Maria"
    When I create root org jabatan "Koordinator Lektor" assigned to umat "Yohanes Budi"
    Then I see org jabatan "Koordinator Lektor" held by "Yohanes Budi"
    When I delete org jabatan "Ketua"
    Then I see org error "Data organisasi masih digunakan oleh periode, sub-organisasi, atau struktur jabatan aktif sehingga tidak dapat dihapus."
    When I delete org period "Periode 2025-2028"
    Then I see org error "Data organisasi masih digunakan oleh periode, sub-organisasi, atau struktur jabatan aktif sehingga tidak dapat dihapus."
    When I delete org jabatan "Bendahara"
    And I delete org jabatan "Dirigen"
    And I delete org jabatan "Koordinator Lektor"
    And I delete org jabatan "Ketua"
    Then I do not see org jabatan "Ketua"
