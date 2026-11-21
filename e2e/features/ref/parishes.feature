Feature: Super Admin parish settings, dioceses, parishes, and global lookups

  @super-admin
  Scenario: Super Admin sees only global settings tiles and is restricted from parish-admin routes
    When I open "/ref"
    Then I see the settings page
    And I see settings tiles "Keuskupan & Paroki, Kamus Referensi"
    And I do not see settings tiles "Profil Paroki, Wilayah & Lingkungan, Periode"
    When I open "/ref/parishes"
    And I create a diocese with code "e2e-kaj-1" and name "Keuskupan E2E 1"
    And I create a parish under diocese "Keuskupan E2E 1" with code "e2e-p-1" and name "Paroki E2E 1"
    And I open "/ref/profile"
    Then I see the parish admin restriction banner
    When I open "/ref/territory"
    Then I see the parish admin restriction banner
    When I open "/ref/periods"
    Then I see the parish admin restriction banner

  @authenticated
  Scenario: Regular user is denied access to dioceses and parishes management
    When I open "/ref/parishes"
    Then I see the super admin restriction banner on parishes page

  @super-admin
  Scenario: Super Admin manages diocese and parish tenant lifecycle
    When I open "/ref/parishes"
    Then I see the parishes page
    When I create a diocese with code "e2e-dioc-del" and name "Keuskupan Hapus E2E"
    Then I see diocese "Keuskupan Hapus E2E" with code "e2e-dioc-del"
    When I create a parish under diocese "Keuskupan Hapus E2E" with code "e2e-par-del" and name "Paroki Hapus E2E"
    Then I see parish "Paroki Hapus E2E" with code "e2e-par-del"
    When I delete parish "Paroki Hapus E2E"
    Then I do not see parish "Paroki Hapus E2E"
    When I delete diocese "Keuskupan Hapus E2E"
    Then I do not see diocese "Keuskupan Hapus E2E"

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
