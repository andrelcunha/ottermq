document.addEventListener('DOMContentLoaded', function() {
    fetchQueues();

    document.getElementById('add-queue-form').addEventListener('submit', function(e) {
        e.preventDefault();
        const queueName = document.getElementById('queue-name').value;
        addQueue(queueName);
    });

    document.getElementById('get-message-form').addEventListener('submit', function(e) {
        e.preventDefault();
        const queue = document.getElementById('selected-queue').value;
        getMessage(queue);
    });
});

async function fetchQueues() {
    const response = await fetch('/api/queues');
    const data = await response.json();
    const queuesList = document.getElementById('queues-list');
    queuesList.innerHTML = '';
    for (const queue of data.queues) {
        const count = await CountMessages(queue)
        const row = document.createElement('tr');
        row.onclick = () => selectQueue(queue);
        row.innerHTML = `
            <td>localhost</td>
            <td><b>${queue}</b></td>
            <td>running</td>
            <td>${count}</td>
            <td>0</td>
            <td>0</td>
            <td>0</td>
            <td>
                <button onclick="deleteQueue('${queue}')">Delete</button>
            </td>
        `;
        queuesList.appendChild(row);
    };
}

async function CountMessages(name) {
    const response = await fetch(`/api/queues/${name}/count`);
    const json = await response.json()
    return json.data.count
}

async function addQueue(name) {
    const response = await fetch('/api/queues', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({name})
    });
    if (response.ok) fetchQueues();
}

async function deleteQueue(name) {
    const response = await fetch(`/api/queues/${name}`, { method: 'DELETE' });
    if (response.ok) fetchQueues();
}

async function getMessage(queue) {
    const response = await fetch(`/api/queues/${queue}/consume`, {
        method: 'POST',
        headers: {'Content-Type': 'application/json'}
    });
    if (response.ok) {
        const json = await response.json();
        const message = json.data.content
        const messageContent = document.getElementById('message-content');
        messageContent.style.display = 'block';
        document.getElementById('message').textContent = message;
        fetchQueues();
    } else {
        console.error('Failed to get message');
    }
}

function selectQueue(queue) {
    document.getElementById('selected-queue').value = queue;
    document.getElementById('get-message').style.display = 'block';
}