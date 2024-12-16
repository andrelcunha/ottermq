document.addEventListener('DOMContentLoaded', function() {
    fetchConnections();
});

async function fetchConnections() {
    const response = await fetch('/api/connections');
    const data = await response.json();
    const connectionsList = document.getElementById('connections-list');
    connectionsList.innerHTML = '';
    for (const connection of data.connections) {
        const row = document.createElement('tr');
        const activeSSL = false
        const sslChar = activeSSL? '●': '○'
        row.onclick = () => selectConnection(connection);
        row.innerHTML = `
            <td>localhost</td>
            <td><b>${connection}</b></td>
            <td>admin</td>
            <td><span class="small-green-square"> </span> running</td>
            <td class='centered'>${sslChar}</td>
            <td class='right'>1</td>
            <td class='right'>-</td>
            <td class='right'>-</td>
        `;
        connectionsList.appendChild(row);
    };
}




function selectConnection(connection) {
    document.getElementById('selected-connection').value = connection;
    // document.getElementById('get-message').style.display = 'block';
}