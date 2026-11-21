Feature: Parish Admin launcher visibility, parish members, and org bootstrap

  @parish-admin
  Scenario: Parish Admin sees all 4 launcher tiles and parish/org scopes on /users
    When I open "/"
    Then I see the home page with admin launcher menus
    When I open home launcher tile "Pengguna"
    Then I am on "/users"
    And I see users directory scopes "Admin & Anggota Paroki, Pengurus & Anggota Organisasi"
    And I do not see users directory scopes "Pengguna Global"

  @parish-admin
  Scenario: Parish Admin manages parish members with last admin demotion guard
    When I open "/users"
    Then I see the users directory page
    When I change parish member "Admin Paroki E2E" role to "Anggota"
    Then I see users error "Minimal harus ada satu Admin Paroki yang tersisa."
    When I appoint parish member with email "umat.paroki.e2e@domus.test" name "Umat Paroki E2E" and role "Anggota"
    Then I see parish member "Umat Paroki E2E" with role "Anggota"
    When I change parish member "Umat Paroki E2E" role to "Admin Paroki"
    Then I see parish member "Umat Paroki E2E" with role "Admin Paroki"
    When I change parish member "Umat Paroki E2E" role to "Anggota"
    Then I see parish member "Umat Paroki E2E" with role "Anggota"
    When I remove parish member "Umat Paroki E2E"
    Then I do not see parish member "Umat Paroki E2E"

  @parish-admin
  Scenario: Parish Admin bootstraps organizations at /org and appoints Org Admin via /users
    Given a zone with code "wil-1-e2e" and name "Wilayah I E2E" exists in the active parish
    When I open "/org"
    Then I see the orgs directory page
    When I create organization with kind "DPP" name "DPP Katedral E2E" code "dpp-e2e"
    And I create organization with kind "Bidang / Seksi" name "Seksi Liturgi E2E" code "liturgi-e2e" parent "DPP Katedral E2E (dpp-e2e)"
    And I create organization with kind "Wilayah / Stasi" name "Pengurus Wilayah I E2E" code "peng-wil-1-e2e" zone "Wilayah I E2E (wil-1-e2e)"
    And I filter organizations by kind "Bidang / Seksi"
    And I search organizations for "Liturgi"
    Then I see organization card "Seksi Liturgi E2E" with code "liturgi-e2e" and kind "Bidang / Seksi"
    When I open "/users"
    And I switch users scope to "Pengurus & Anggota Organisasi"
    And I select organization "Seksi Liturgi E2E"
    And I self-assign as Organization Admin
    Then I see org member "Admin Paroki E2E" with role "Admin Organisasi"
    When I add org member with email "ketua.liturgi.e2e@domus.test" name "Ketua Liturgi E2E" and role "Admin Organisasi"
    Then I see org member "Ketua Liturgi E2E" with role "Admin Organisasi"
