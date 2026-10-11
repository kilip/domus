Feature: Login

  Scenario: Guest is redirected to login
    When I open "/"
    Then I should be on "/login"
    And I should see the login form

  Scenario: Login page shows OAuth providers
    When I open "/login"
    Then I should see the login form
