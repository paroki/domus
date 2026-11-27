Feature: Umat administration, Kartu Keluarga, 3-step Umat wizard, and Sakramen

  @paroki-admin
  Scenario: Paroki Admin delegates umat:admin in /users and non-umat:admin is denied /umat
    Given a paroki member "umat-admin-e2e@domus.test" named "Admin Umat E2E" exists in the active paroki
    When I open "/umat"
    Then I see the umat forbidden card
    When I open "/users"
    And I open user detail for "umat-admin-e2e@domus.test"
    And I switch to user detail tab "Paroki"
    And I grant umat admin for the active paroki
    Then I see the umat admin badge for the active paroki
    When I sign in as "umat-admin-e2e@domus.test" on the active paroki
    And I open "/umat"
    Then I see the umat directory page

  @paroki-admin
  Scenario: Admin Data Umat manages Kartu Keluarga, Umat wizard, Sakramen, and delete guards
    Given I am signed in as umat admin on the active paroki
    When I open "/umat/keluarga"
    And I create a Kartu Keluarga named "Keluarga E2E" with nomorKk "3174010101908888" and alamat "Jl. Katedral 1"
    Then I see decrypted nomorKk "3174010101908888" on keluarga detail
    When I open "/umat/keluarga"
    And I search keluarga by nomorKk "3174010101908888"
    Then I see masked keluarga "Keluarga E2E" with "************8888" and no raw "3174010101908888"
    When I create an umat via wizard named "Yohanes E2E" with nik "3174010101907777" linked to "Keluarga E2E"
    And I edit umat nama baptis to "Yohanes Pembaptis"
    And I link umat to paroki member "umat-e2e@domus.test"
    And I add sakramen "Baptis" on "2000-05-10" buku "B-01" nomor "100/2000"
    And I edit sakramen "Baptis" nomor to "101/2000"
    And I delete sakramen "Baptis"
    When I search umat by nik "3174010101907777"
    Then I see masked umat "Yohanes E2E" with "************7777" and no raw "3174010101907777"
    When I open "/umat/keluarga"
    And I open keluarga detail for "Keluarga E2E"
    And I try to delete the Kartu Keluarga
    Then I see umat error "Kartu Keluarga masih memiliki anggota umat aktif dan tidak dapat dihapus."
    When I open "/umat"
    And I open umat detail for "Yohanes E2E"
    And I delete the umat
    And I open "/umat/keluarga"
    And I open keluarga detail for "Keluarga E2E"
    And I delete the Kartu Keluarga
    Then I am on "/umat/keluarga"

  @paroki-admin
  Scenario: Admin Data Umat manages Kamus Referensi Umat, /umat/new picker, UMAT_IN_USE, and plain user read-only
    Given I am signed in as umat admin on the active paroki
    When I create a paroki umat referensi in category "agama" with code "katolik-e2e" and label "Katolik E2E"
    Then I see umat referensi "Katolik E2E" with scope "Paroki"
    And the "Agama" picker on "/umat/new" includes "Katolik E2E"
    When I create an umat via wizard named "Budi Ref E2E" with nik "3174010101906666" and agama "Katolik E2E"
    And I open "/umat/referensi?category=agama"
    And I try to delete umat referensi "Katolik E2E"
    Then I see umat error "Item referensi masih digunakan oleh data umat atau keluarga dan tidak dapat dihapus."
    When I open "/umat"
    And I open umat detail for "Budi Ref E2E"
    And I delete the umat
    And I open "/umat/referensi?category=agama"
    And I edit umat referensi "Katolik E2E" to label "Katolik E2E Updated"
    Then I see umat referensi "Katolik E2E Updated" with scope "Paroki"
    When I toggle active state for umat referensi "Katolik E2E Updated"
    And I delete umat referensi "Katolik E2E Updated"
    Then I do not see umat referensi "Katolik E2E Updated"
    Given a paroki member "umat-ro-e2e@domus.test" named "Member RO E2E" exists in the active paroki
    When I sign in as "umat-ro-e2e@domus.test" on the active paroki
    And I open "/umat/referensi?category=agama"
    Then umat referensi list is read-only without write controls
