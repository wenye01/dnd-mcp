// Package store_test contains integration tests for the message store
package store_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/dnd-mcp/server/internal/models"
	"github.com/dnd-mcp/server/internal/store/postgres"
	"github.com/dnd-mcp/server/pkg/config"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getMessageTestModuleRoot returns the module root directory
func getMessageTestModuleRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(filename)
	return filepath.Dir(filepath.Dir(filepath.Dir(testDir)))
}

// getMessageTestMigrationsPath returns the absolute path to migrations directory
func getMessageTestMigrationsPath() string {
	return filepath.Join(getMessageTestModuleRoot(), "internal", "store", "postgres", "migrations")
}

func getMessageTestConfig() config.PostgresConfig {
	return config.PostgresConfig{
		Host:            getMessageTestEnvOrDefault("POSTGRES_HOST", "localhost"),
		Port:            5432,
		User:            getMessageTestEnvOrDefault("POSTGRES_USER", "postgres"),
		Password:        getMessageTestEnvOrDefault("POSTGRES_PASSWORD", "password"),
		DBName:          getMessageTestEnvOrDefault("POSTGRES_TEST_DBNAME", "dnd_server_test"),
		SSLMode:         "disable",
		PoolSize:        5,
		MaxConnLifetime: 3600,
		MaxConnIdleTime: 1800,
	}
}

func getMessageTestEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// setupMessageTestDB sets up the test database and returns a client and cleanup function
func setupMessageTestDB(t *testing.T) (*postgres.Client, *postgres.CampaignStore, *postgres.MessageStore, func()) {
	t.Helper()

	// Load .env file
	envPath := filepath.Join(getMessageTestModuleRoot(), ".env")
	if _, err := os.Stat(envPath); err == nil {
		godotenv.Load(envPath)
	}

	cfg := getMessageTestConfig()

	client, err := postgres.NewClient(cfg)
	require.NoError(t, err, "Failed to create client")

	// Check if database is available
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		client.Close()
		t.Skipf("Database not available: %v", err)
	}

	// Run migrations
	migrator := postgres.NewMigratorWithPath(client, getMessageTestMigrationsPath())
	err = migrator.Up(ctx)
	require.NoError(t, err, "Failed to run migrations")

	campaignStore := postgres.NewCampaignStore(client)
	messageStore := postgres.NewMessageStore(client)

	cleanup := func() {
		// Clean up test data
		ctx := context.Background()
		client.Pool().Exec(ctx, "DELETE FROM messages")
		client.Pool().Exec(ctx, "DELETE FROM campaigns")
		client.Close()
	}

	return client, campaignStore, messageStore, cleanup
}

func TestMessageStore_Create(t *testing.T) {
	_, campaignStore, messageStore, cleanup := setupMessageTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign first
	campaign := models.NewCampaign("Message Test Campaign", "dm-message-001", "")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err, "Failed to create campaign")

	// Create a user message
	message := models.NewUserMessage(campaign.ID, "player-001", "Hello, world!")
	err = messageStore.Create(ctx, message)
	require.NoError(t, err, "Failed to create message")

	assert.NotEmpty(t, message.ID, "ID should be set")
	assert.NotZero(t, message.CreatedAt, "CreatedAt should be set")
	assert.Equal(t, models.MessageRoleUser, message.Role)
	assert.Equal(t, "player-001", message.PlayerID)
}

func TestMessageStore_Create_AssistantMessage(t *testing.T) {
	_, campaignStore, messageStore, cleanup := setupMessageTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign first
	campaign := models.NewCampaign("Assistant Message Test", "dm-message-002", "")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err)

	// Create an assistant message with tool calls
	toolCalls := []models.ToolCall{
		{
			ID:   "call-001",
			Name: "roll_dice",
			Arguments: map[string]interface{}{
				"sides":  20,
				"count":  1,
			},
		},
	}
	message := models.NewAssistantMessage(campaign.ID, "I rolled the dice for you.", toolCalls)

	err = messageStore.Create(ctx, message)
	require.NoError(t, err, "Failed to create assistant message")

	// Verify retrieval
	retrieved, err := messageStore.Get(ctx, message.ID)
	require.NoError(t, err, "Failed to get message")

	assert.Equal(t, models.MessageRoleAssistant, retrieved.Role)
	assert.Len(t, retrieved.ToolCalls, 1)
	assert.Equal(t, "roll_dice", retrieved.ToolCalls[0].Name)
}

func TestMessageStore_Create_SystemMessage(t *testing.T) {
	_, campaignStore, messageStore, cleanup := setupMessageTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign first
	campaign := models.NewCampaign("System Message Test", "dm-message-003", "")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err)

	// Create a system message
	message := models.NewMessage(campaign.ID, models.MessageRoleSystem, "You are a helpful D&D assistant.")

	err = messageStore.Create(ctx, message)
	require.NoError(t, err, "Failed to create system message")

	// Verify retrieval
	retrieved, err := messageStore.Get(ctx, message.ID)
	require.NoError(t, err)

	assert.Equal(t, models.MessageRoleSystem, retrieved.Role)
	assert.Empty(t, retrieved.PlayerID)
}

func TestMessageStore_Get(t *testing.T) {
	_, campaignStore, messageStore, cleanup := setupMessageTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign and message
	campaign := models.NewCampaign("Get Message Test", "dm-message-004", "")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err)

	message := models.NewUserMessage(campaign.ID, "player-002", "Test message content")
	err = messageStore.Create(ctx, message)
	require.NoError(t, err)

	// Get the message
	retrieved, err := messageStore.Get(ctx, message.ID)
	require.NoError(t, err, "Failed to get message")

	assert.Equal(t, message.ID, retrieved.ID)
	assert.Equal(t, campaign.ID, retrieved.CampaignID)
	assert.Equal(t, models.MessageRoleUser, retrieved.Role)
	assert.Equal(t, "Test message content", retrieved.Content)
	assert.Equal(t, "player-002", retrieved.PlayerID)
}

func TestMessageStore_Get_NotFound(t *testing.T) {
	_, _, messageStore, cleanup := setupMessageTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := messageStore.Get(ctx, uuid.New().String())
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound")
}

func TestMessageStore_GetByCampaignID(t *testing.T) {
	_, campaignStore, messageStore, cleanup := setupMessageTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create campaigns and messages
	campaign1 := models.NewCampaign("Campaign 1", "dm-message-005", "")
	err := campaignStore.Create(ctx, campaign1)
	require.NoError(t, err)

	campaign2 := models.NewCampaign("Campaign 2", "dm-message-006", "")
	err = campaignStore.Create(ctx, campaign2)
	require.NoError(t, err)

	// Create messages in campaign1
	msg1 := models.NewUserMessage(campaign1.ID, "player-003", "Message in campaign 1")
	err = messageStore.Create(ctx, msg1)
	require.NoError(t, err)

	// Create message in campaign2 with same content
	msg2 := models.NewUserMessage(campaign2.ID, "player-004", "Message in campaign 1")
	err = messageStore.Create(ctx, msg2)
	require.NoError(t, err)

	// Get message by campaign ID
	retrieved, err := messageStore.GetByCampaignID(ctx, campaign1.ID, msg1.ID)
	require.NoError(t, err)

	assert.Equal(t, msg1.ID, retrieved.ID)
	assert.Equal(t, campaign1.ID, retrieved.CampaignID)

	// Try to get campaign2's message using campaign1's ID
	_, err = messageStore.GetByCampaignID(ctx, campaign1.ID, msg2.ID)
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound")
}

func TestMessageStore_ListByCampaign(t *testing.T) {
	_, campaignStore, messageStore, cleanup := setupMessageTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign
	campaign := models.NewCampaign("List Messages Test", "dm-message-007", "")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err)

	// Create multiple messages
	messages := []*models.Message{
		models.NewMessage(campaign.ID, models.MessageRoleSystem, "System init"),
		models.NewUserMessage(campaign.ID, "player-005", "First user message"),
		models.NewAssistantMessage(campaign.ID, "Assistant response", nil),
		models.NewUserMessage(campaign.ID, "player-005", "Second user message"),
	}

	for _, msg := range messages {
		err = messageStore.Create(ctx, msg)
		require.NoError(t, err)
		// Small delay to ensure different timestamps
		time.Sleep(1 * time.Millisecond)
	}

	// List all messages
	retrieved, err := messageStore.ListByCampaign(ctx, campaign.ID, 0)
	require.NoError(t, err, "Failed to list messages")
	assert.Len(t, retrieved, 4, "Should have 4 messages")

	// Verify order (ascending by created_at)
	assert.Equal(t, messages[0].ID, retrieved[0].ID)
	assert.Equal(t, messages[3].ID, retrieved[3].ID)
}

func TestMessageStore_ListByCampaign_WithLimit(t *testing.T) {
	_, campaignStore, messageStore, cleanup := setupMessageTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign
	campaign := models.NewCampaign("Limit Test", "dm-message-008", "")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err)

	// Create multiple messages
	for i := 0; i < 10; i++ {
		msg := models.NewUserMessage(campaign.ID, "player-006", fmt.Sprintf("Message %d", i))
		err = messageStore.Create(ctx, msg)
		require.NoError(t, err)
		time.Sleep(1 * time.Millisecond)
	}

	// List with limit
	retrieved, err := messageStore.ListByCampaign(ctx, campaign.ID, 5)
	require.NoError(t, err)
	assert.Len(t, retrieved, 5, "Should return 5 messages")
}

func TestMessageStore_ListByCampaignWithOffset(t *testing.T) {
	_, campaignStore, messageStore, cleanup := setupMessageTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign
	campaign := models.NewCampaign("Offset Test", "dm-message-009", "")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err)

	// Create multiple messages
	var messageIDs []string
	for i := 0; i < 10; i++ {
		msg := models.NewUserMessage(campaign.ID, "player-007", fmt.Sprintf("Message %d", i))
		err = messageStore.Create(ctx, msg)
		require.NoError(t, err)
		messageIDs = append(messageIDs, msg.ID)
		time.Sleep(1 * time.Millisecond)
	}

	// Get first page
	page1, err := messageStore.ListByCampaignWithOffset(ctx, campaign.ID, 5, 0)
	require.NoError(t, err)
	assert.Len(t, page1, 5, "First page should have 5 messages")
	assert.Equal(t, messageIDs[0], page1[0].ID)

	// Get second page
	page2, err := messageStore.ListByCampaignWithOffset(ctx, campaign.ID, 5, 5)
	require.NoError(t, err)
	assert.Len(t, page2, 5, "Second page should have 5 messages")
	assert.Equal(t, messageIDs[5], page2[0].ID)
}

func TestMessageStore_CountByCampaign(t *testing.T) {
	_, campaignStore, messageStore, cleanup := setupMessageTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign
	campaign := models.NewCampaign("Count Test", "dm-message-010", "")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err)

	// Count before messages
	count, err := messageStore.CountByCampaign(ctx, campaign.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "Should have 0 messages initially")

	// Create messages
	for i := 0; i < 7; i++ {
		msg := models.NewUserMessage(campaign.ID, "player-008", fmt.Sprintf("Message %d", i))
		err = messageStore.Create(ctx, msg)
		require.NoError(t, err)
	}

	// Count after messages
	count, err = messageStore.CountByCampaign(ctx, campaign.ID)
	require.NoError(t, err)
	assert.Equal(t, 7, count, "Should have 7 messages")
}

func TestMessageStore_Delete(t *testing.T) {
	_, campaignStore, messageStore, cleanup := setupMessageTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign and message
	campaign := models.NewCampaign("Delete Test", "dm-message-011", "")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err)

	message := models.NewUserMessage(campaign.ID, "player-009", "To be deleted")
	err = messageStore.Create(ctx, message)
	require.NoError(t, err)

	// Delete message
	err = messageStore.Delete(ctx, message.ID)
	require.NoError(t, err, "Failed to delete message")

	// Verify deletion
	_, err = messageStore.Get(ctx, message.ID)
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound after delete")
}

func TestMessageStore_Delete_NotFound(t *testing.T) {
	_, _, messageStore, cleanup := setupMessageTestDB(t)
	defer cleanup()

	ctx := context.Background()

	err := messageStore.Delete(ctx, uuid.New().String())
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Should return ErrNotFound")
}

func TestMessageStore_DeleteByCampaign(t *testing.T) {
	_, campaignStore, messageStore, cleanup := setupMessageTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create campaigns
	campaign1 := models.NewCampaign("Campaign Delete 1", "dm-message-012", "")
	err := campaignStore.Create(ctx, campaign1)
	require.NoError(t, err)

	campaign2 := models.NewCampaign("Campaign Delete 2", "dm-message-013", "")
	err = campaignStore.Create(ctx, campaign2)
	require.NoError(t, err)

	// Create messages in both campaigns
	msg1 := models.NewUserMessage(campaign1.ID, "player-010", "Message 1")
	err = messageStore.Create(ctx, msg1)
	require.NoError(t, err)

	msg2 := models.NewUserMessage(campaign2.ID, "player-010", "Message 2")
	err = messageStore.Create(ctx, msg2)
	require.NoError(t, err)

	// Delete all messages from campaign1
	err = messageStore.DeleteByCampaign(ctx, campaign1.ID)
	require.NoError(t, err, "Failed to delete campaign messages")

	// Verify campaign1 messages are deleted
	_, err = messageStore.Get(ctx, msg1.ID)
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Campaign1 message should be deleted")

	// Verify campaign2 messages still exist
	retrieved, err := messageStore.Get(ctx, msg2.ID)
	require.NoError(t, err, "Campaign2 message should still exist")
	assert.Equal(t, msg2.ID, retrieved.ID)
}

func TestMessageStore_DeleteByCampaignBeforeDate(t *testing.T) {
	_, campaignStore, messageStore, cleanup := setupMessageTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign
	campaign := models.NewCampaign("Date Delete Test", "dm-message-014", "")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err)

	// Create old message (simulate by manually setting created_at)
	oldMessage := models.NewUserMessage(campaign.ID, "player-011", "Old message")
	oldMessage.CreatedAt = time.Now().Add(-48 * time.Hour) // 2 days ago
	err = messageStore.Create(ctx, oldMessage)
	require.NoError(t, err)

	// Create recent message
	recentMessage := models.NewUserMessage(campaign.ID, "player-011", "Recent message")
	err = messageStore.Create(ctx, recentMessage)
	require.NoError(t, err)

	// Delete messages older than 24 hours
	cutoff := time.Now().Add(-24 * time.Hour)
	deleted, err := messageStore.DeleteByCampaignBeforeDate(ctx, campaign.ID, cutoff)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, deleted, int64(1), "Should delete at least 1 old message")

	// Verify old message is deleted
	_, err = messageStore.Get(ctx, oldMessage.ID)
	assert.ErrorIs(t, err, postgres.ErrNotFound, "Old message should be deleted")

	// Verify recent message still exists
	retrieved, err := messageStore.Get(ctx, recentMessage.ID)
	require.NoError(t, err, "Recent message should still exist")
	assert.Equal(t, recentMessage.ID, retrieved.ID)
}

func TestMessageStore_ToolCalls_JSON(t *testing.T) {
	_, campaignStore, messageStore, cleanup := setupMessageTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a campaign
	campaign := models.NewCampaign("Tool Calls Test", "dm-message-015", "")
	err := campaignStore.Create(ctx, campaign)
	require.NoError(t, err)

	// Create message with complex tool calls
	toolCalls := []models.ToolCall{
		{
			ID:   "call-001",
			Name: "roll_dice",
			Arguments: map[string]interface{}{
				"sides":      int(20),
				"count":      int(2),
				"modifier":   int(5),
				"advantage":  true,
				"character":  "Fighter",
			},
			Result: &models.ToolResult{
				Success: true,
				Data: map[string]interface{}{
					"rolls":    []int{15, 8},
					"total":    28,
					"modifier": 5,
				},
			},
		},
		{
			ID:   "call-002",
			Name: "check_condition",
			Arguments: map[string]interface{}{
				"condition": "prone",
				"character": "Goblin",
			},
			Result: &models.ToolResult{
				Success: false,
				Error:   "Character not found",
			},
		},
	}

	message := models.NewAssistantMessage(campaign.ID, "I performed some checks.", toolCalls)
	err = messageStore.Create(ctx, message)
	require.NoError(t, err)

	// Retrieve and verify
	retrieved, err := messageStore.Get(ctx, message.ID)
	require.NoError(t, err)

	assert.Len(t, retrieved.ToolCalls, 2)

	// Verify first tool call
	tc1 := retrieved.ToolCalls[0]
	assert.Equal(t, "call-001", tc1.ID)
	assert.Equal(t, "roll_dice", tc1.Name)
	assert.Equal(t, int(20), int(tc1.Arguments["sides"].(float64)))
	assert.True(t, tc1.Result.Success)
	assert.Equal(t, int(28), int(tc1.Result.Data["total"].(float64)))

	// Verify second tool call
	tc2 := retrieved.ToolCalls[1]
	assert.Equal(t, "call-002", tc2.ID)
	assert.Equal(t, "check_condition", tc2.Name)
	assert.False(t, tc2.Result.Success)
	assert.Equal(t, "Character not found", tc2.Result.Error)
}
