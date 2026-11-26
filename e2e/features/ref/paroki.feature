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

  @super-admin
  Scenario: Super Admin manages kevikepan, sipil hierarchy with REF_IN_USE, and paroki wizard
    When I create a keuskupan with code "e2e-kaj-wiz" and name "Keuskupan Wizard E2E"
    And I create a kevikepan under keuskupan "Keuskupan Wizard E2E" with code "pusat-e2e" and name "Kevikepan Pusat E2E"
    Then I see kevikepan "Kevikepan Pusat E2E" with code "pusat-e2e"
    When I create a kota with code "3171-e2e" and name "Kota Jakarta Pusat E2E"
    And I create a kecamatan under kota "Kota Jakarta Pusat E2E" with code "317101-e2e" and name "Gambir E2E"
    And I create a kelurahan under kecamatan "Gambir E2E" with code "31710101-e2e" name "Gambir Kel E2E" and kodePos "10110"
    When I delete kota "Kota Jakarta Pusat E2E"
    Then I see ref error "Data masih digunakan oleh data lain dan tidak dapat dihapus."
    When I delete kelurahan "Gambir Kel E2E"
    Then I do not see kelurahan "Gambir Kel E2E"
    When I delete kecamatan "Gambir E2E"
    Then I do not see kecamatan "Gambir E2E"
    When I delete kota "Kota Jakarta Pusat E2E"
    Then I do not see kota "Kota Jakarta Pusat E2E"
    When I create a paroki under keuskupan "Keuskupan Wizard E2E" with code "e2e-p-wiz" and name "Paroki Wizard E2E"
    And I edit paroki struktur kodeResmi to "KAJ-WIZ-01"
    Then I see paroki "Paroki Wizard E2E" with code "e2e-p-wiz"
    When I delete paroki "Paroki Wizard E2E"
    Then I do not see paroki "Paroki Wizard E2E"
    When I delete kevikepan "Kevikepan Pusat E2E"
    Then I do not see kevikepan "Kevikepan Pusat E2E"
    When I delete keuskupan "Keuskupan Wizard E2E"
    Then I do not see keuskupan "Keuskupan Wizard E2E"
