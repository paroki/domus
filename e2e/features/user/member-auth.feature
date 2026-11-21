Feature: Invite-only login guard, guest invite sign-up, onboarding, and profile security

  Scenario: Uninvited email is rejected on /login
    When I open "/login"
    Then I see the login page
    When I request email OTP on login for "stranger.uninvited@domus.test"
    Then I see login error containing "Akun tidak ditemukan"

  Scenario: New guest signs up via /org/join/:code, completes onboarding, and lands on /org/:orgId
    Given an organization "Paduan Suara E2E" has an active member invite with position "Anggota Paduan Suara"
    When I open the organization invite link
    Then I see invite preview for "Paduan Suara E2E" with role "Anggota" and position "Anggota Paduan Suara"
    When I request guest OTP with name "Tamu Undangan E2E" and email "tamu.undangan.e2e@domus.test"
    And a deterministic OTP "123456" is seeded for "tamu.undangan.e2e@domus.test"
    And I verify guest OTP "123456" for "Paduan Suara E2E"
    Then I am redirected to onboarding
    And I see onboarding birth date bounds
    When I complete onboarding with name "Tamu Undangan Terverifikasi", phone "+6281298765432", gender "wanita", and birth date "1996-08-17"
    Then I am on the invited organization detail page "Paduan Suara E2E"
    And I see read-only org roster with member "Tamu Undangan Terverifikasi"
    When I open "/me"
    Then I see the profile and security page
    And I see my profile shows gender "wanita" and birth date "1996-08-17"
    When I update my profile to name "Tamu Undangan Terverifikasi", phone "+6281298765432", gender "pria", and birth date "1988-04-12"
    Then I see profile saved message
    And I see my profile shows gender "pria" and birth date "1988-04-12"

  Scenario: Revoked invite link is rejected on /org/join/:code
    Given an organization "Paduan Suara Revoked E2E" has a revoked invite
    When I open the organization invite link
    Then I see revoked invite message "Tautan undangan ini berstatus Dicabut dan tidak dapat digunakan lagi."

  @authenticated
  Scenario: Authenticated user updates profile, revokes secondary session, verifies last account guard, and logs out
    Given a secondary session exists for "user@domus.test"
    When I open "/me"
    Then I see the profile and security page
    When I update my profile to name "Umat Mandiri E2E", phone "+6281399998888", gender "pria", and birth date "1992-03-14"
    Then I see profile saved message
    And I see my profile shows gender "pria" and birth date "1992-03-14"
    And I see current device session in active sessions
    When I revoke the secondary session
    Then I see last linked account protection
    When I log out from header menu
    Then I am on "/login"
