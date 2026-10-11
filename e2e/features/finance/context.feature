Feature: Finance Unit Context

  Scenario: Automatically select first accessible unit when accessing finance
    Given I am logged in as a regular user
    And I have finance access to unit "E2E Unit Keuangan A" in parish "E2E Paroki Keuangan"
    And I have finance access to unit "E2E Unit Keuangan B" in parish "E2E Paroki Keuangan"
    When I clear finance local storage
    And I open "/fin"
    Then I should be on "/fin"
    And I should see active finance unit "E2E Unit Keuangan A"
    And localStorage activeFinanceUnitId should be set for unit "E2E Unit Keuangan A"

  Scenario: Switch active finance unit via selector
    Given I am logged in as a regular user
    And I have finance access to unit "E2E Unit Keuangan A" in parish "E2E Paroki Keuangan"
    And I have finance access to unit "E2E Unit Keuangan B" in parish "E2E Paroki Keuangan"
    When I clear finance local storage
    And I open "/fin"
    Then I should see active finance unit "E2E Unit Keuangan A"
    When I switch active finance unit to "E2E Unit Keuangan B"
    Then I should see active finance unit "E2E Unit Keuangan B"
    And localStorage activeFinanceUnitId should be set for unit "E2E Unit Keuangan B"

  Scenario: User with no finance access sees warning
    Given I am logged in as a regular user
    When I clear finance local storage
    And I open "/fin"
    Then I should be on "/fin"
    And I should see finance no access warning

