Feature: Workspace CRUD
  As an admin
  I want to manage workspaces in the system
  So that parish organizational units and groups are properly maintained

  Scenario: Create a new workspace
    Given I am logged in as an admin
    And parish "E2E Paroki Workspace Test" exists
    And I open "/sys/workspace"
    When I click the add workspace button
    Then I should be on "/sys/workspace/create"
    When I fill the workspace name with "E2E Dewan Paroki Harian"
    And I select parish "E2E Paroki Workspace Test" in workspace form
    And I fill the workspace description with "Dewan paroki harian periode aktif"
    And I submit the workspace form
    Then I should be on "/sys/workspace"
    And I should see a success notification "Workspace ditambahkan."
    And I should see workspace "E2E Dewan Paroki Harian" in the list
    And I should see workspace "E2E Dewan Paroki Harian" with status "Aktif" in the list

  Scenario: Workspace form validation
    Given I am logged in as an admin
    And I open "/sys/workspace"
    When I click the add workspace button
    Then I should be on "/sys/workspace/create"
    When I submit the workspace form
    Then I should see a workspace validation error "Nama workspace wajib diisi."
    And I should see a workspace validation error "Paroki wajib dipilih."
    When I cancel the workspace form
    Then I should be on "/sys/workspace"

  Scenario: Edit an existing workspace
    Given I am logged in as an admin
    And parish "E2E Paroki Workspace Edit" exists
    And workspace "E2E Workspace To Edit" exists in parish "E2E Paroki Workspace Edit"
    And I open "/sys/workspace"
    When I click edit for workspace "E2E Workspace To Edit"
    And I fill the workspace name with "E2E Workspace Edited"
    And I submit the workspace form
    Then I should be on "/sys/workspace"
    And I should see a success notification "Workspace diperbarui."
    And I should see workspace "E2E Workspace Edited" in the list
    And I should not see workspace "E2E Workspace To Edit" in the list

  Scenario: Delete a workspace
    Given I am logged in as an admin
    And parish "E2E Paroki Workspace Delete" exists
    And workspace "E2E Workspace To Delete" exists in parish "E2E Paroki Workspace Delete"
    And I open "/sys/workspace"
    When I click delete for workspace "E2E Workspace To Delete"
    And I confirm the deletion
    Then I should see a success notification "Workspace dihapus."
    And I should not see workspace "E2E Workspace To Delete" in the list

  Scenario: Search workspace by name
    Given I am logged in as an admin
    And parish "E2E Paroki Workspace Search" exists
    And workspace "E2E Workspace Alfa" exists in parish "E2E Paroki Workspace Search"
    And workspace "E2E Workspace Beta" exists in parish "E2E Paroki Workspace Search"
    And I open "/sys/workspace"
    When I search workspace for "Alfa"
    Then I should see workspace "E2E Workspace Alfa" in the list
    And I should not see workspace "E2E Workspace Beta" in the list

