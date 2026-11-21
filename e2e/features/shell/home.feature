Feature: Home page smoke

  Scenario: Anonymous visitor is redirected to login
    When I open "/"
    Then I am on "/login"
    And I see the login page

  @super-admin
  Scenario: Super Admin sees home page and launcher menus
    When I open "/"
    Then I am on "/"
    And I see the home page with admin launcher menus
