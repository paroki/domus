Feature: Paroki Admin launcher visibility, paroki members, and org bootstrap

  @paroki-admin
  Scenario: Paroki Admin sees all 4 launcher tiles and paroki role filters on /users
    When I open "/"
    Then I see the home page with admin launcher menus
    When I open home launcher tile "Pengguna"
    Then I am on "/users"
    And I see users filter buttons "Semua Peran, Admin Paroki, Anggota"
    And I do not see users directory controls "Pengguna Global, Semua status"

  @paroki-admin
  Scenario: Paroki Admin manages paroki members with last admin demotion guard
    Given a paroki member with email "umat.paroki.e2e@domus.test" name "Umat Paroki E2E" exists in the active paroki
    When I open "/users"
    Then I see the users directory page
    When I open user detail "Admin Paroki E2E"
    And I demote paroki membership in the active paroki
    Then I see users error "Minimal harus ada satu Admin Paroki yang tersisa."
    When I go back to the users directory
    And I open user detail "Umat Paroki E2E"
    And I promote paroki membership in the active paroki
    Then I see active paroki membership with role "Admin Paroki"
    When I demote paroki membership in the active paroki
    Then I see active paroki membership with role "Anggota"
    When I remove paroki membership in the active paroki
    And I go back to the users directory
    Then I do not see user "umat.paroki.e2e@domus.test"

  @paroki-admin
  Scenario: Paroki Admin bootstraps organizations at /org and appoints Org Admin via /users
    Given a wilayah with code "wil-1-e2e" and name "Wilayah I E2E" exists in the active paroki
    And a paroki member with email "ketua.liturgi.e2e@domus.test" name "Ketua Liturgi E2E" exists in the active paroki
    When I open "/org"
    Then I see the orgs directory page
    When I create organization with kind "DPP" name "DPP Katedral E2E" code "dpp-e2e"
    And I create organization with kind "Bidang / Seksi" name "Seksi Liturgi E2E" code "liturgi-e2e" parent "DPP Katedral E2E (dpp-e2e)"
    And I create organization with kind "Wilayah / Stasi" name "Pengurus Wilayah I E2E" code "peng-wil-1-e2e" wilayah "Wilayah I E2E (wil-1-e2e)"
    And I filter organizations by kind "Bidang / Seksi"
    And I search organizations for "Liturgi"
    Then I see organization card "Seksi Liturgi E2E" with code "liturgi-e2e" and kind "Bidang / Seksi"
    When I open "/users"
    And I open user detail "Admin Paroki E2E"
    And I switch user detail tab to "Organisasi"
    And I add user to org "Seksi Liturgi E2E" with role "Admin Organisasi"
    Then I see user org membership "Seksi Liturgi E2E" with role "Admin Organisasi"
    When I go back to the users directory
    And I open user detail "Ketua Liturgi E2E"
    And I switch user detail tab to "Organisasi"
    And I add user to org "Seksi Liturgi E2E" with role "Admin Organisasi"
    Then I see user org membership "Seksi Liturgi E2E" with role "Admin Organisasi"
