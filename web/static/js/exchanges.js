document.addEventListener('DOMContentLoaded', function() {
    fetchExchanges();

    document.getElementById('add-exchange-form').addEventListener('submit', function(e) {
        e.preventDefault();
        const exchangeName = document.getElementById('exchange-name').value;
        addExchange(exchangeName);
    });

    document.getElementById('add-binding-form').addEventListener('submit', function(e) {
        e.preventDefault();
        const exchange = document.getElementById('selected-exchange').value;
        const routingKey = document.getElementById('routing-key').value;
        const queue = document.getElementById('queue-name').value;
        addBinding(exchange, routingKey, queue);
    });

    document.getElementById('publish-message-form').addEventListener('submit', function(e) {
        e.preventDefault();
        const exchange = document.getElementById('selected-exchange-for-message').value;
        const routingKey = document.getElementById('msg-routing-key').value;
        const message = document.getElementById('payload').value;
        publishMessage(exchange, routingKey, message);
    });
});

async function fetchExchanges() {
    const response = await fetch('/api/exchanges');
    const data = await response.json();
    const exchangesList = document.getElementById('exchanges-list');
    exchangesList.innerHTML = '';
    data.exchanges.forEach(exchange => {
        const row = document.createElement('tr');
        row.onclick = () => selectExchange(exchange);
        row.innerHTML = `
            <td style="display:none;">${exchange.vhost_id}</td>
            <td>${exchange.vhost}</td>
            <td><b>${exchange.name}<b></td>
            <td>${exchange.type}</td>
            <td>
                <button class='delete-button' onclick="deleteExchange('${exchange.vhost_id}','${exchange.name}')">Delete</button>
            </td>
        `;
        exchangesList.appendChild(row);
    });
}

async function addExchange(name, vhost_id) {
    const exchange = {
        exchange_name: name,
        exchange_type: "direct",
        // vhost_id: vhost_id
    }
    const response = await fetch('/api/exchanges', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(exchange)
    });
    if (response.ok) fetchExchanges();
}

async function deleteExchange(vhost_id, name) {
    const response = await fetch(`/api/exchanges/${vhost_id}/${name}`, { method: 'DELETE' });
    if (response.ok) fetchExchanges();
}

async function fetchBindings(vhost_id, exchange) {
    const response = await fetch(`/api/bindings/${vhost_id}/${exchange}`);
    const data = await response.json();
    const bindingsList = document.getElementById('bindings-list');
    bindingsList.innerHTML = '';
    Object.entries(data.bindings).forEach(([routingKey, queues]) => {
        queues.forEach(queue => {
            const row = document.createElement('tr');
            row.innerHTML = `
                <td>${routingKey}</td>
                <td>${queue}</td>
                <td>
                    <button onclick="deleteBinding('${exchange}', '${routingKey}', '${queue}')">Unbind</button>
                </td>
            `;
            bindingsList.appendChild(row);
        });
    });
}

async function addBinding(vhost_id, exchange, routingKey, queue) {
    const binding = {
        vhost_name: vhost,
        exchange_name: exchange,
        routing_key: routingKey,
        queue_name: queue,
    }
    const response = await fetch('/api/bindings', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(binding)
    });
    if (response.ok) fetchBindings(vhost_id, exchange);
}

async function deleteBinding(vhost_id, exchange, routingKey, queue) {
    const binding = {
        exchange_name: exchange,
        routing_key: routingKey,
        queue_name: queue,
    }
    const response = await fetch('/api/bindings', {
        method: 'DELETE',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(binding)
    });
    if (response.ok) fetchBindings(vhost_id, exchange);
}

async function publishMessage(exchange, routingKey, message) {
    const response = await fetch('/api/messages', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({exchange_name: exchange, routing_key: routingKey, message: message})
    });
    if (response.ok) alert('Message published successfully!');
}

function selectExchange(exchange) {
    document.getElementById('selected-exchange').value = exchange;
    document.getElementById('selected-exchange-for-message').value = exchange;
    document.getElementById('bindings-manager').style.display = 'block';
    document.getElementById('publish-message').style.display = 'block';
    fetchBindings(exchange.vhost_id, exchange.name);
}