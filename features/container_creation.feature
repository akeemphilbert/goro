Feature: Container Creation and Hierarchy Management
  As a pod user
  I want to organize my data in folders
  So that I can logically group related resources

  Background:
    Given a clean LDP server is running
    And the server supports container operations

  Scenario: Create a basic container
    When I create a container with ID "documents"
    Then the container should be created successfully
    And the container should have type "BasicContainer"
    And the container should be empty

  Scenario: Create nested container hierarchy
    Given a container "documents" exists
    When I create a container "images" inside "documents"
    Then the container "images" should be created successfully
    And the container "images" should have parent "documents"
    And the container "documents" should contain "images"

  Scenario: Create deep hierarchy
    Given a container "root" exists
    And a container "level1" exists inside "root"
    When I create a container "level2" inside "level1"
    Then the container "level2" should be created successfully
    And the hierarchy path should be "root/level1/level2"

  Scenario: Prevent circular references
    Given a container "parent" exists
    And a container "child" exists inside "parent"
    When I try to move container "parent" inside "child"
    Then the operation should fail with "circular reference" error

  Scenario: Container creation with metadata
    When I create a container with ID "photos" and title "Photo Collection"
    Then the container should be created successfully
    And the container should have title "Photo Collection"
    And the container should have creation timestamp
    And the container should have Dublin Core metadata

  Scenario: Container creation validation
    When I try to create a container with invalid ID ""
    Then the operation should fail with "invalid container ID" error

  Scenario: Duplicate container prevention
    Given a container "documents" exists
    When I try to create another container with ID "documents"
    Then the operation should fail with "container already exists" error