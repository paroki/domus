Feature: Super Admin paroki settings, keuskupan, paroki, and global lookups

  @super-admin
  Scenario: Super Admin sees only global settings tiles and is restricted from paroki-admin routes
    When I open "/ref"
    Then I see the settings page
    And I see settings tiles "Keuskupan & Paroki, Kamus Referensi"
    And I do not see settings tiles "Profil Paroki, Wilayah & Lingkungan, Periode"
    When I open "/ref/paroki"
    And I create a keuskupan with code "e2e-kaj-1" and name "Keuskupan E2E 1"
    And I create a paroki under keuskupan "Keuskupan E2E 1" with code "e2e-p-1" and name "Paroki E2E 1"
    And I open "/ref/profile"
    Then I see the paroki admin restriction banner
    When I open "/ref/wilayah"
    Then I see the paroki admin restriction banner
    When I open "/ref/periods"
    Then I see the paroki admin restriction banner

  @authenticated
  Scenario: Regular user is denied access to keuskupan and paroki management
    When I open "/ref/paroki"
    Then I see the super admin restriction banner on paroki page

  @super-admin
  Scenario: Super Admin manages keuskupan and paroki tenant lifecycle
    When I open "/ref/paroki"
    Then I see the paroki page
    When I create a keuskupan with code "e2e-dioc-del" and name "Keuskupan Hapus E2E"
    Then I see keuskupan "Keuskupan Hapus E2E" with code "e2e-dioc-del"
    When I create a paroki under keuskupan "Keuskupan Hapus E2E" with code "e2e-par-del" and name "Paroki Hapus E2E"
    Then I see paroki "Paroki Hapus E2E" with code "e2e-par-del"
    When I delete paroki "Paroki Hapus E2E"
    Then I do not see paroki "Paroki Hapus E2E"
    When I delete keuskupan "Keuskupan Hapus E2E"
    Then I do not see keuskupan "Keuskupan Hapus E2E"

  @super-admin
  Scenario: Super Admin performs global reference lookups CRUD
    When I open "/ref/lookups"
    Then I see the lookups page
    When I create a global lookup in category "sacrament" with code "e2e-baptis" and label "Baptis E2E"
    Then I see lookup "Baptis E2E" with scope "Global"
    When I edit lookup "Baptis E2E" to label "Baptis E2E Updated"
    Then I see lookup "Baptis E2E Updated" with scope "Global"
    When I toggle active state for lookup "Baptis E2E Updated"
    And I delete lookup "Baptis E2E Updated"
    Then I do not see lookup "Baptis E2E Updated"
