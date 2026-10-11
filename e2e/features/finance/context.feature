Feature: Finance Workspace Context

  Scenario: Automatically select first accessible workspace when accessing finance
    Given I am logged in as a regular user
    And I have finance access to workspace "E2E Workspace Keuangan A" in parish "E2E Paroki Keuangan"
    And I have finance access to workspace "E2E Workspace Keuangan B" in parish "E2E Paroki Keuangan"
    When I clear finance local storage
    And I open "/fin"
    Then I should be on "/fin"
    And I should see active finance workspace "E2E Workspace Keuangan A"
    And localStorage activeFinanceWorkspaceId should be set for workspace "E2E Workspace Keuangan A"

  Scenario: Switch active finance workspace via selector
    Given I am logged in as a regular user
    And I have finance access to workspace "E2E Workspace Keuangan A" in parish "E2E Paroki Keuangan"
    And I have finance access to workspace "E2E Workspace Keuangan B" in parish "E2E Paroki Keuangan"
    When I clear finance local storage
    And I open "/fin"
    Then I should see active finance workspace "E2E Workspace Keuangan A"
    When I switch active finance workspace to "E2E Workspace Keuangan B"
    Then I should see active finance workspace "E2E Workspace Keuangan B"
    And localStorage activeFinanceWorkspaceId should be set for workspace "E2E Workspace Keuangan B"

  Scenario: User with no finance access sees warning
    Given I am logged in as a regular user
    When I clear finance local storage
    And I open "/fin"
    Then I should be on "/fin"
    And I should see finance no access warning

