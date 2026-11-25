Feature: Super Admin paroki settings, keuskupan, paroki, and global lookups

  @super-admin
  Scenario: Super Admin sees only global settings tiles and is restricted from paroki-admin routes
    When I open "/settings"
    Then I see the settings page
    And I see settings tiles "Struktur Gerejawi, Wilayah Sipil, Kamus Referensi, Sistem & Integrasi"
    And I do not see settings tiles "Administrasi Paroki"
    When I open "/settings/gerejawi/paroki"
    And I create a keuskupan with code "e2e-kaj-1" and name "Keuskupan E2E 1"
    And I create a paroki under keuskupan "Keuskupan E2E 1" with code "e2e-p-1" and name "Paroki E2E 1"
    And I open "/settings/paroki/profile"
    Then I see the paroki admin restriction banner
    When I open "/settings/paroki/wilayah"
    Then I see the paroki admin restriction banner
    When I open "/settings/paroki/periods"
    Then I see the paroki admin restriction banner

  @authenticated
  Scenario: Regular user is denied access to keuskupan and paroki management
    When I open "/settings/gerejawi/paroki"
    Then I see the super admin restriction banner on paroki page

  @super-admin
  Scenario: Super Admin manages keuskupan and paroki tenant lifecycle
    When I open "/settings/gerejawi/paroki"
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
    When I open "/settings/lookups"
    Then I see the lookups page
    When I create a global lookup in category "status_paroki" with code "e2e-mandiri" and label "Mandiri E2E"
    Then I see lookup "Mandiri E2E" with scope "Global"
    When I edit lookup "Mandiri E2E" to label "Mandiri E2E Updated"
    Then I see lookup "Mandiri E2E Updated" with scope "Global"
    When I toggle active state for lookup "Mandiri E2E Updated"
    And I delete lookup "Mandiri E2E Updated"
    Then I do not see lookup "Mandiri E2E Updated"
