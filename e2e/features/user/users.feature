Feature: Super Admin users directory and parish admin appointment

  @authenticated
  Scenario: Non-admin user is denied access to users directory
    When I open "/users"
    Then I see the users forbidden banner

  @super-admin
  Scenario: Super Admin searches, promotes, disables, and deletes a global user
    Given a global user with email "petugas.e2e@domus.test" name "Petugas Paroki E2E" phone "+6281234567890" exists
    When I open "/users"
    Then I see the users directory page
    When I search users for "petugas.e2e@domus.test"
    Then I see global user "Petugas Paroki E2E" with email "petugas.e2e@domus.test" and status "Aktif"
    When I open user detail "petugas.e2e@domus.test"
    And I promote the user to Super Admin
    And I demote the user from Super Admin
    And I disable the user
    And I go back to the users directory
    And I filter users by status "Nonaktif"
    And I search users for "petugas.e2e@domus.test"
    Then I see global user "Petugas Paroki E2E" with email "petugas.e2e@domus.test" and status "Nonaktif"
    When I open user detail "petugas.e2e@domus.test"
    And I enable the user
    And I go back to the users directory
    And I filter users by status "Semua status"
    And I search users for "petugas.e2e@domus.test"
    Then I see global user "Petugas Paroki E2E" with email "petugas.e2e@domus.test" and status "Aktif"
    When I open user detail "petugas.e2e@domus.test"
    And I delete the user
    And I search users for "petugas.e2e@domus.test"
    Then I do not see user "petugas.e2e@domus.test"

  @super-admin
  Scenario: Super Admin appoints a Parish Admin via user detail Paroki tab
    Given a diocese "Keuskupan Users E2E" and parish "Paroki Users E2E" exist
    And a global user with email "parishadmin.e2e@domus.test" name "Admin Paroki E2E" phone "+6281234567891" exists
    When I open "/users"
    And I search users for "parishadmin.e2e@domus.test"
    And I open user detail "parishadmin.e2e@domus.test"
    And I switch user detail tab to "Paroki"
    And I add user to parish "Paroki Users E2E" with role "Admin Paroki"
    Then I see user parish membership "Paroki Users E2E" with role "Admin Paroki"
