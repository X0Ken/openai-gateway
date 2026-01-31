package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler handles web UI requests
type Handler struct{}

// NewHandler creates a new web handler
func NewHandler() *Handler {
	return &Handler{}
}

// RegisterRoutes registers web routes
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	// Serve index page
	r.GET("/", h.Index)
}

// Index serves the main admin page
func (h *Handler) Index(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, `<!DOCTYPE html>
<html>
<head>
    <title>OpenAI Gateway Admin</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .section { margin: 20px 0; padding: 20px; border: 1px solid #ddd; }
        table { width: 100%; border-collapse: collapse; }
        th, td { padding: 10px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #f5f5f5; }
        button { padding: 8px 16px; margin: 5px; cursor: pointer; }
        .enabled { color: green; }
        .disabled { color: red; }
    </style>
</head>
<body>
    <h1>OpenAI Gateway Admin</h1>
    
    <div class="section">
        <h2>Models</h2>
        <div id="models"></div>
        <button onclick="loadModels()">Refresh</button>
    </div>
    
    <div class="section">
        <h2>Channels</h2>
        <div id="channels"></div>
        <button onclick="loadChannels()">Refresh</button>
    </div>
    
    <div class="section">
        <h2>Users</h2>
        <div id="users"></div>
        <button onclick="loadUsers()">Refresh</button>
    </div>
    
    <div class="section">
        <h2>Sessions</h2>
        <div id="sessions"></div>
        <button onclick="loadSessions()">Refresh</button>
    </div>
    
    <script>
        async function loadModels() {
            const resp = await fetch('/api/models');
            const models = await resp.json();
            document.getElementById('models').innerHTML = renderModels(models);
        }
        
        async function loadChannels() {
            const resp = await fetch('/api/channels');
            const channels = await resp.json();
            document.getElementById('channels').innerHTML = renderChannels(channels);
        }
        
        async function loadUsers() {
            const resp = await fetch('/api/users');
            const users = await resp.json();
            document.getElementById('users').innerHTML = renderUsers(users);
        }
        
        async function loadSessions() {
            const resp = await fetch('/api/sessions');
            const sessions = await resp.json();
            document.getElementById('sessions').innerHTML = renderSessions(sessions);
        }
        
        function renderModels(models) {
            if (!models || models.length === 0) return '<p>No models configured</p>';
            let html = '<table><tr><th>ID</th><th>Name</th><th>Channels</th></tr>';
            models.forEach(m => {
                html += '<tr><td>' + m.id + '</td><td>' + m.name + '</td><td>' + m.channels_count + ' channel(s)</td></tr>';
            });
            html += '</table>';
            return html;
        }
        
        function renderChannels(channels) {
            if (!channels || channels.length === 0) return '<p>No channels configured</p>';
            let html = '<table><tr><th>ID</th><th>Name</th><th>Base URL</th><th>Weight</th><th>Status</th></tr>';
            channels.forEach(ch => {
                const status = ch.enabled ? '<span class="enabled">Enabled</span>' : '<span class="disabled">Disabled</span>';
                html += '<tr><td>' + ch.id + '</td><td>' + ch.name + '</td><td>' + ch.base_url + '</td><td>' + ch.weight + '</td><td>' + status + '</td></tr>';
            });
            html += '</table>';
            return html;
        }
        
        function renderUsers(users) {
            if (!users || users.length === 0) return '<p>No users configured</p>';
            let html = '<table><tr><th>ID</th><th>Name</th><th>API Key</th></tr>';
            users.forEach(u => {
                html += '<tr><td>' + u.id + '</td><td>' + (u.name || '') + '</td><td>' + u.api_key.substring(0, 10) + '...</td></tr>';
            });
            html += '</table>';
            return html;
        }
        
        function renderSessions(sessions) {
            if (!sessions || sessions.length === 0) return '<p>No active sessions</p>';
            let html = '<table><tr><th>ID</th><th>User ID</th><th>Channel ID</th><th>Last Used</th></tr>';
            sessions.forEach(s => {
                html += '<tr><td>' + s.id + '</td><td>' + s.user_id + '</td><td>' + s.channel_id + '</td><td>' + s.last_used_at + '</td></tr>';
            });
            html += '</table>';
            return html;
        }
        
        // Load data on page load
        loadModels();
        loadChannels();
        loadUsers();
        loadSessions();
    </script>
</body>
</html>`)
}
