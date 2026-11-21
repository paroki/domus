Feature: Super Admin users directory and parish admin appointment

  @authenticated
  Scenario: Non-admin user is denied access to users directory
    When I open "/users"
    Then I see the users forbidden banner

  @super-admin
  Scenario: Super Admin creates, searches, promotes, disables, and deletes a global user
    When I open "/users"
    Then I see the users directory page
    When I create a global user with email "petugas.e2e@domus.test" name "Petugas Paroki E2E" phone "+6281234567890"
    And I search users for "petugas.e2e@domus.test"
    Then I see global user "Petugas Paroki E2E" with email "petugas.e2e@domus.test" and status "Aktif"
    When I promote user "petugas.e2e@domus.test" to Super Admin
    Then I see global user "Petugas Paroki E2E" with Super Admin badge
    When I demote user "petugas.e2e@domus.test" from Super Admin
    Then I see global user "Petugas Paroki E2E" without Super Admin badge
    When I disable user "petugas.e2e@domus.test"
    And I filter users by status "Nonaktif"
    Then I see global user "Petugas Paroki E2E" with email "petugas.e2e@domus.test" and status "Nonaktif"
    When I enable user "petugas.e2e@domus.test"
    And I filter users by status "Semua status"
    Then I see global user "Petugas Paroki E2E" with email "petugas.e2e@domus.test" and status "Aktif"
    When I delete global user "Petugas Paroki E2E"
    Then I do not see user "petugas.e2e@domus.test"

  @super-admin
  Scenario: Super Admin appoints a Parish Admin via Admin & Anggota Paroki scope
    Given a diocese "Keuskupan Users E2E" and parish "Paroki Users E2E" exist
    When I open "/users"
    And I switch users scope to "Admin & Anggota Paroki"
    And I select parish "Paroki Users E2E"
    And I appoint parish member with email "parishadmin.e2e@domus.test" name "Admin Paroki E2E" and role "Admin Paroki"
    Then I see parish member "Admin Paroki E2E" with role "Admin Paroki"
