document.addEventListener('DOMContentLoaded', function() {
    fetchConnections();
});

async function fetchConnections() {
    const response = await fetch('/api/connections');
    const data = await response.json();
    const connectionsList = document.getElementById('connections-list');
    connectionsList.innerHTML = '';
    for (const connInfo of data.connections) {
        const heartbeat_in_seconds = getLastHeatbeatInSecs(connInfo.last_heartbeat)

        const date = new Date(connInfo.connected_at)
        const fmt_time = formatTime(date)
        const fmt_date = formatDate(date)

        const row = document.createElement('tr');
        const fmt_ssl = formatIsSSL(false);
        row.onclick = () => selectConnection(connInfo);
        row.innerHTML = `
            <td>localhost</td>
            <td><b>${connInfo.name}</b></td>
            <td>admin</td>
            <td><span class="small-green-square"> </span> running</td>
            <td class='centered'>${fmt_ssl}</td>
            <td class='right'>1</td>
            <td class='right'>${heartbeat_in_seconds}s</td>
            <td class='right'><span class='show-time'>${fmt_time}</span></br><span class='show-date'>${fmt_date}</span></td>
        `;
        connectionsList.appendChild(row);
    };
}

function formatIsSSL(activeSSL) {
    const sslChar = activeSSL ? '●' : '○';
    return sslChar;
}

function getLastHeatbeatInSecs(last_heartbeat) {
    const lastHeartbeatDate = new Date(last_heartbeat)
    const now = new Date()
    const diffMs = now.getTime() - lastHeartbeatDate.getTime()
    const diff = moment.duration(diffMs, 'milliseconds');
    const seconds = Math.floor(diff.asSeconds());
    return seconds;
}

function formatTime(date) {
    const hours = date.getHours().toString().padStart(2, '0')
    const minutes = date.getMinutes().toString().padStart(2, '0')
    const seconds = date.getSeconds().toString().padStart(2, '0')
    return `${hours}:${minutes}:${seconds}`
}

function formatDate(date) {
    const year = date.getFullYear();
    const month = (date.getMonth() + 1).toString().padStart(2, '0')
    const day = date.getDate().toString().padStart(2, '0')
    return `${year}-${month}-${day}`
}

function selectConnection(connection) {
    document.getElementById('selected-connection').value = connection;
}

setInterval(fetchConnections, 15000)