Feature: Account Management and Membership
  As a pod owner and account administrator
  I want to manage accounts and user memberships
  So that I can control access and organize users within my pod

  Background:
    Given a clean user management system is running
    And the system supports account operations
    And system roles are seeded
    And users exist for testing

  Scenario: Account creation with owner assignment
    Given a user exists with ID "owner123" and email "owner@example.com"
    When I create an account with name "My Pod Account" and owner "owner123"
    Then the account should be created successfully
    And the account should have name "My Pod Account"
    And the account should have owner "owner123"
    And the owner should be automatically added as a member with "owner" role
    And an AccountCreatedEvent should be emitted
    And a MemberAddedEvent should be emitted

  Scenario: Account creation with invalid owner fails
    When I try to create an account with name "Test Account" and owner "nonexistent"
    Then the creation should fail with "owner not found" error
    And no AccountCreatedEvent should be emitted

  Scenario: Account creation with empty name fails
    Given a user exists with ID "owner123"
    When I try to create an account with name "" and owner "owner123"
    Then the creation should fail with "account name is required" error
    And no account should be created

  Scenario: User invitation with role assignment
    Given an account exists with ID "account123" and owner "owner123"
    And a user exists with ID "inviter123" with "admin" role in account "account123"
    When I invite user "newuser@example.com" to account "account123" with role "member" by user "inviter123"
    Then the invitation should be created successfully
    And the invitation should have status "pending"
    And the invitation should have role "member"
    And the invitation should have a unique token
    And the invitation should expire in 7 days
    And a MemberInvitedEvent should be emitted

  Scenario: User invitation with invalid role fails
    Given an account exists with ID "account123" and owner "owner123"
    When I try to invite user "test@example.com" to account "account123" with role "invalid" by user "owner123"
    Then the invitation should fail with "invalid role" error
    And no MemberInvitedEvent should be emitted

  Scenario: User invitation by non-admin fails
    Given an account exists with ID "account123" and owner "owner123"
    And a user exists with ID "member123" with "member" role in account "account123"
    When I try to invite user "test@example.com" to account "account123" with role "member" by user "member123"
    Then the invitation should fail with "insufficient permissions" error
    And no invitation should be created

  Scenario: Invitation acceptance creates membership
    Given an account exists with ID "account123"
    And an invitation exists with token "invite123" for email "newuser@example.com" with role "member"
    And a user exists with ID "user456" and email "newuser@example.com"
    When I accept invitation with token "invite123" for user "user456"
    Then the invitation should be accepted successfully
    And the invitation status should be "accepted"
    And the user should be added as a member with role "member"
    And an InvitationAcceptedEvent should be emitted
    And a MemberAddedEvent should be emitted

  Scenario: Invitation acceptance with invalid token fails
    Given a user exists with ID "user456"
    When I try to accept invitation with token "invalid" for user "user456"
    Then the acceptance should fail with "invitation not found" error
    And no InvitationAcceptedEvent should be emitted

  Scenario: Invitation acceptance with expired token fails
    Given an invitation exists with token "expired123" that has expired
    And a user exists with ID "user456"
    When I try to accept invitation with token "expired123" for user "user456"
    Then the acceptance should fail with "invitation expired" error
    And no membership should be created

  Scenario: Invitation acceptance with wrong email fails
    Given an invitation exists with token "invite123" for email "correct@example.com"
    And a user exists with ID "user456" and email "wrong@example.com"
    When I try to accept invitation with token "invite123" for user "user456"
    Then the acceptance should fail with "email mismatch" error
    And no membership should be created

  Scenario: Member role update by admin
    Given an account exists with ID "account123" and owner "owner123"
    And a user exists with ID "member456" with "member" role in account "account123"
    When I update member "member456" role to "admin" in account "account123" by user "owner123"
    Then the role update should succeed
    And the member should have role "admin"
    And a MemberRoleUpdatedEvent should be emitted

  Scenario: Member role update by non-admin fails
    Given an account exists with ID "account123" and owner "owner123"
    And a user exists with ID "member456" with "member" role in account "account123"
    And a user exists with ID "member789" with "member" role in account "account123"
    When I try to update member "member456" role to "admin" in account "account123" by user "member789"
    Then the role update should fail with "insufficient permissions" error
    And no MemberRoleUpdatedEvent should be emitted

  Scenario: Member role update to invalid role fails
    Given an account exists with ID "account123" and owner "owner123"
    And a user exists with ID "member456" with "member" role in account "account123"
    When I try to update member "member456" role to "invalid" in account "account123" by user "owner123"
    Then the role update should fail with "invalid role" error
    And the member role should remain unchanged

  Scenario: Member removal by admin
    Given an account exists with ID "account123" and owner "owner123"
    And a user exists with ID "member456" with "member" role in account "account123"
    When I remove member "member456" from account "account123" by user "owner123"
    Then the member should be removed successfully
    And the user should no longer be a member of the account
    And a MemberRemovedEvent should be emitted

  Scenario: Member removal by non-admin fails
    Given an account exists with ID "account123" and owner "owner123"
    And a user exists with ID "member456" with "member" role in account "account123"
    And a user exists with ID "member789" with "member" role in account "account123"
    When I try to remove member "member456" from account "account123" by user "member789"
    Then the removal should fail with "insufficient permissions" error
    And the member should remain in the account

  Scenario: Owner cannot be removed from account
    Given an account exists with ID "account123" and owner "owner123"
    When I try to remove member "owner123" from account "account123" by user "owner123"
    Then the removal should fail with "cannot remove account owner" error
    And the owner should remain in the account

  Scenario: Account membership listing
    Given an account exists with ID "account123" and owner "owner123"
    And users exist with roles in account "account123":
      | user_id   | role   |
      | admin456  | admin  |
      | member789 | member |
      | viewer012 | viewer |
    When I list members of account "account123"
    Then the listing should return 4 members
    And the listing should include owner "owner123" with role "owner"
    And the listing should include user "admin456" with role "admin"
    And the listing should include user "member789" with role "member"
    And the listing should include user "viewer012" with role "viewer"

  Scenario: Account membership listing with permissions
    Given an account exists with ID "account123" and owner "owner123"
    And a user exists with ID "member456" with "member" role in account "account123"
    When I list members of account "account123" as user "member456"
    Then the listing should succeed
    And the listing should show member information based on role permissions

  Scenario: Account membership listing by non-member fails
    Given an account exists with ID "account123" and owner "owner123"
    And a user exists with ID "outsider456" not in any account
    When I try to list members of account "account123" as user "outsider456"
    Then the listing should fail with "access denied" error

  Scenario: Role-based permission validation for resource access
    Given an account exists with ID "account123" and owner "owner123"
    And a user exists with ID "viewer456" with "viewer" role in account "account123"
    When I check if user "viewer456" can "create" resources in account "account123"
    Then the permission check should return false
    When I check if user "viewer456" can "read" resources in account "account123"
    Then the permission check should return true

  Scenario: Role-based permission validation for user management
    Given an account exists with ID "account123" and owner "owner123"
    And a user exists with ID "admin456" with "admin" role in account "account123"
    When I check if user "admin456" can "invite" users to account "account123"
    Then the permission check should return true
    When I check if user "admin456" can "delete" the account "account123"
    Then the permission check should return false

  Scenario: Account settings management
    Given an account exists with ID "account123" and owner "owner123"
    When I update account "account123" settings with max_members "50" and allow_invitations "true"
    Then the account settings should be updated successfully
    And the account should have max_members "50"
    And the account should have allow_invitations "true"

  Scenario: Account settings update by non-owner fails
    Given an account exists with ID "account123" and owner "owner123"
    And a user exists with ID "admin456" with "admin" role in account "account123"
    When I try to update account "account123" settings by user "admin456"
    Then the update should fail with "only owner can modify account settings" error

  Scenario: Invitation expiration handling
    Given an account exists with ID "account123"
    And an invitation exists with token "expire123" that will expire in 1 second
    When I wait for 2 seconds
    And I try to accept invitation with token "expire123" for any user
    Then the acceptance should fail with "invitation expired" error
    And the invitation status should be "expired"

  Scenario: Multiple invitations to same email
    Given an account exists with ID "account123" and owner "owner123"
    And an invitation exists for email "test@example.com" with status "pending"
    When I try to invite user "test@example.com" to account "account123" again
    Then the invitation should fail with "invitation already pending" error
    And no duplicate invitation should be created

  Scenario: Invitation revocation by admin
    Given an account exists with ID "account123" and owner "owner123"
    And an invitation exists with token "revoke123" for email "test@example.com" with status "pending"
    When I revoke invitation with token "revoke123" by user "owner123"
    Then the invitation should be revoked successfully
    And the invitation status should be "revoked"
    And the invitation cannot be accepted anymore

  Scenario: Account deletion with member cleanup
    Given an account exists with ID "account123" and owner "owner123"
    And users exist with roles in account "account123":
      | user_id   | role   |
      | member456 | member |
      | admin789  | admin  |
    When I delete account "account123" by owner "owner123"
    Then the account should be deleted successfully
    And all memberships should be removed
    And all pending invitations should be revoked
    And an AccountDeletedEvent should be emitted

  Scenario: Account deletion by non-owner fails
    Given an account exists with ID "account123" and owner "owner123"
    And a user exists with ID "admin456" with "admin" role in account "account123"
    When I try to delete account "account123" by user "admin456"
    Then the deletion should fail with "only owner can delete account" error
    And the account should remain active

  Scenario: Concurrent membership operations
    Given an account exists with ID "account123" and owner "owner123"
    When multiple clients simultaneously try to add different members to account "account123"
    Then all valid operations should succeed
    And no duplicate memberships should be created
    And all MemberAddedEvents should be emitted correctly

  Scenario: Concurrent invitation operations
    Given an account exists with ID "account123" and owner "owner123"
    When multiple clients simultaneously try to invite different users to account "account123"
    Then all valid invitations should be created
    And each invitation should have a unique token
    And all MemberInvitedEvents should be emitted correctly

  Scenario: Account membership limits enforcement
    Given an account exists with ID "account123" with max_members "3"
    And the account already has 3 members including owner
    When I try to invite another user to account "account123"
    Then the invitation should fail with "account member limit reached" error
    And no invitation should be created

  Scenario: Atomic account operations with event consistency
    Given an account exists with ID "account123" and owner "owner123"
    When I perform multiple membership operations in sequence:
      | operation | user_id   | role   |
      | add       | member1   | member |
      | add       | admin1    | admin  |
      | update    | member1   | admin  |
      | remove    | admin1    |        |
    Then all operations should complete successfully
    And the final membership state should be consistent
    And all corresponding events should be emitted in correct order