Feature: Parish Admin launcher visibility, parish members, and org bootstrap

  @parish-admin
  Scenario: Parish Admin sees all 4 launcher tiles and parish role filters on /users
    When I open "/"
    Then I see the home page with admin launcher menus
    When I open home launcher tile "Pengguna"
    Then I am on "/users"
    And I see users filter buttons "Semua Peran, Admin Paroki, Anggota"
    And I do not see users directory controls "Pengguna Global, Semua status"

  @parish-admin
  Scenario: Parish Admin manages parish members with last admin demotion guard
    Given a parish member with email "umat.paroki.e2e@domus.test" name "Umat Paroki E2E" exists in the active parish
    When I open "/users"
    Then I see the users directory page
    When I open user detail "Admin Paroki E2E"
    And I demote parish membership in the active parish
    Then I see users error "Minimal harus ada satu Admin Paroki yang tersisa."
    When I go back to the users directory
    And I open user detail "Umat Paroki E2E"
    And I promote parish membership in the active parish
    Then I see active parish membership with role "Admin Paroki"
    When I demote parish membership in the active parish
    Then I see active parish membership with role "Anggota"
    When I remove parish membership in the active parish
    And I go back to the users directory
    Then I do not see user "umat.paroki.e2e@domus.test"

  @parish-admin
  Scenario: Parish Admin bootstraps organizations at /org and appoints Org Admin via /users
    Given a zone with code "wil-1-e2e" and name "Wilayah I E2E" exists in the active parish
    And a parish member with email "ketua.liturgi.e2e@domus.test" name "Ketua Liturgi E2E" exists in the active parish
    When I open "/org"
    Then I see the orgs directory page
    When I create organization with kind "DPP" name "DPP Katedral E2E" code "dpp-e2e"
    And I create organization with kind "Bidang / Seksi" name "Seksi Liturgi E2E" code "liturgi-e2e" parent "DPP Katedral E2E (dpp-e2e)"
    And I create organization with kind "Wilayah / Stasi" name "Pengurus Wilayah I E2E" code "peng-wil-1-e2e" zone "Wilayah I E2E (wil-1-e2e)"
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
