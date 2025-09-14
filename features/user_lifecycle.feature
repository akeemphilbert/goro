Feature: User Lifecycle Management
  As a pod administrator and user
  I want to manage user accounts throughout their lifecycle
  So that I can control access and users can manage their own data

  Background:
    Given a clean user management system is running
    And the system supports user operations
    And system roles are seeded

  Scenario: Successful user registration with WebID creation
    When I register a new user with email "john.doe@example.com" and name "John Doe"
    Then the user should be created successfully
    And a unique WebID should be generated for the user
    And the WebID document should be created in Turtle format
    And the user should have status "active"
    And a UserRegisteredEvent should be emitted
    And a WebIDGeneratedEvent should be emitted

  Scenario: User registration with duplicate email fails
    Given a user exists with email "existing@example.com"
    When I try to register a new user with email "existing@example.com"
    Then the registration should fail with "user already exists" error
    And no UserRegisteredEvent should be emitted

  Scenario: User registration with invalid email fails
    When I try to register a new user with email "invalid-email"
    Then the registration should fail with "invalid email format" error
    And no user should be created

  Scenario: User registration with empty name fails
    When I try to register a new user with email "test@example.com" and name ""
    Then the registration should fail with "name is required" error
    And no user should be created

  Scenario: User profile update
    Given a user exists with ID "user123" and email "john@example.com"
    When I update the user profile with name "John Smith" and bio "Software Developer"
    Then the user profile should be updated successfully
    And the user name should be "John Smith"
    And the user bio should be "Software Developer"
    And a UserProfileUpdatedEvent should be emitted

  Scenario: User profile update with invalid data fails
    Given a user exists with ID "user123"
    When I try to update the user profile with name ""
    Then the update should fail with "name cannot be empty" error
    And no UserProfileUpdatedEvent should be emitted

  Scenario: User retrieval by ID
    Given a user exists with ID "user123" and email "john@example.com"
    When I retrieve the user by ID "user123"
    Then the user should be returned successfully
    And the user email should be "john@example.com"

  Scenario: User retrieval by WebID
    Given a user exists with WebID "https://pod.example.com/users/john#me"
    When I retrieve the user by WebID "https://pod.example.com/users/john#me"
    Then the user should be returned successfully
    And the user WebID should be "https://pod.example.com/users/john#me"

  Scenario: User retrieval by email
    Given a user exists with email "john@example.com"
    When I retrieve the user by email "john@example.com"
    Then the user should be returned successfully
    And the user email should be "john@example.com"

  Scenario: User retrieval with non-existent ID fails
    When I try to retrieve a user by ID "nonexistent"
    Then the retrieval should fail with "user not found" error

  Scenario: User self-deletion with confirmation
    Given a user exists with ID "user123" and email "john@example.com"
    When I delete the user account with ID "user123"
    Then the user should be deleted successfully
    And the user status should be "deleted"
    And all user files should be cleaned up
    And a UserDeletedEvent should be emitted

  Scenario: User deletion of non-existent user fails
    When I try to delete a user with ID "nonexistent"
    Then the deletion should fail with "user not found" error
    And no UserDeletedEvent should be emitted

  Scenario: User status transitions
    Given a user exists with ID "user123" and status "active"
    When I suspend the user with ID "user123"
    Then the user status should be "suspended"
    When I reactivate the user with ID "user123"
    Then the user status should be "active"

  Scenario: WebID document generation and validation
    When I register a new user with email "webid.test@example.com"
    Then a WebID document should be generated
    And the WebID document should be in valid Turtle format
    And the WebID document should contain the user's name and email
    And the WebID document should contain proper RDF triples

  Scenario: WebID uniqueness validation
    Given a user exists with WebID "https://pod.example.com/users/john#me"
    When I try to create another user with the same WebID
    Then the creation should fail with "WebID already exists" error

  Scenario: User profile preferences management
    Given a user exists with ID "user123"
    When I update the user preferences with theme "dark" and language "en"
    Then the user preferences should be updated successfully
    And the user theme should be "dark"
    And the user language should be "en"

  Scenario: User data export before deletion
    Given a user exists with ID "user123" with profile data and preferences
    When I request user data export for user "user123"
    Then the user data should be exported successfully
    And the export should contain profile information
    And the export should contain WebID document
    And the export should contain user preferences

  Scenario: Concurrent user registration
    When multiple clients simultaneously try to register users with different emails
    Then all registrations should succeed
    And each user should have a unique WebID
    And all UserRegisteredEvents should be emitted

  Scenario: User validation rules enforcement
    When I try to register a user with email longer than 255 characters
    Then the registration should fail with "email too long" error
    When I try to register a user with name longer than 255 characters
    Then the registration should fail with "name too long" error