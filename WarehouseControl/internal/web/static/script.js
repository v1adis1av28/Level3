let jwtToken = localStorage.getItem('token');
let currentUser = localStorage.getItem('user');
let currentRole = localStorage.getItem('role');

if (jwtToken) {
    document.getElementById('login-form').style.display = 'none';
    document.getElementById('main-content').style.display = 'block';
    document.getElementById('current-user').textContent = currentUser;
    document.getElementById('current-role').textContent = currentRole;

    if (currentRole === 'admin' || currentRole === 'manager') {
        document.getElementById('add-item-form').style.display = 'block';
    }

    loadItems();
}

function login() {
    const username = document.getElementById('username').value;
    const role = document.getElementById('role').value;

    fetch('/auth/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ username, role })
    })
    .then(response => {
        // Получаем токен из заголовка Authorization
        const token = response.headers.get('Authorization')?.split(' ')[1];
        return response.json().then(data => ({ data, token }));
    })
    .then(({ data, token }) => {
        if (data.result) {
            // Если токен получен из заголовка, сохраняем его
            if (token) {
                localStorage.setItem('token', token);
            } else {
                // Иначе пробуем достать из localStorage (если был сохранён вручную)
                const storedToken = localStorage.getItem('token');
                if (!storedToken) {
                    alert('Токен не получен');
                    return;
                }
            }

            localStorage.setItem('user', username);
            localStorage.setItem('role', role);
            jwtToken = token || localStorage.getItem('token');
            currentUser = username;
            currentRole = role;

            document.getElementById('login-form').style.display = 'none';
            document.getElementById('main-content').style.display = 'block';
            document.getElementById('current-user').textContent = username;
            document.getElementById('current-role').textContent = role;

            if (currentRole === 'admin' || currentRole === 'manager') {
                document.getElementById('add-item-form').style.display = 'block';
            }

            loadItems();
        } else {
            alert('Ошибка входа: ' + data.error);
        }
    })
    .catch(err => console.error(err));
}

function logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    localStorage.removeItem('role');
    jwtToken = null;
    currentUser = null;
    currentRole = null;

    document.getElementById('login-form').style.display = 'block';
    document.getElementById('main-content').style.display = 'none';
    document.getElementById('items-body').innerHTML = '';
    document.getElementById('history-section').style.display = 'none';
}

function loadItems() {
    fetch('/items', {
        headers: {
            'Authorization': 'Bearer ' + jwtToken
        }
    })
    .then(response => response.json())
    .then(data => {
        if (data.items) {
            const tbody = document.getElementById('items-body');
            tbody.innerHTML = '';

            data.items.forEach(item => {
                const row = document.createElement('tr');
                row.innerHTML = `
                    <td>${item.id}</td>
                    <td>${item.name}</td>
                    <td>${item.description}</td>
                    <td>${item.quantity}</td>
                    <td>
                        ${currentRole === 'admin' || currentRole === 'manager' ? `<button onclick="editItem(${item.id})">Редактировать</button>` : ''}
                        ${currentRole === 'admin' ? `<button onclick="deleteItem(${item.id})">Удалить</button>` : ''}
                        <button onclick="loadHistory(${item.id})">История</button>
                    </td>
                `;
                tbody.appendChild(row);
            });
        }
    })
    .catch(err => console.error(err));
}

function createItem() {
    const name = document.getElementById('item-name').value;
    const description = document.getElementById('item-description').value;
    const quantity = parseInt(document.getElementById('item-quantity').value);

    fetch('/items', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + jwtToken
        },
        body: JSON.stringify({ name, description, quantity })
    })
    .then(response => response.json())
    .then(data => {
        if (data.result) {
            alert('Товар добавлен');
            document.getElementById('item-name').value = '';
            document.getElementById('item-description').value = '';
            document.getElementById('item-quantity').value = '';
            loadItems();
        } else {
            alert('Ошибка: ' + data.error);
        }
    })
    .catch(err => console.error(err));
}

function editItem(id) {
    const name = prompt('Новое название:');
    const description = prompt('Новое описание:');
    const quantity = parseInt(prompt('Новое количество:'));

    fetch(`/items/${id}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + jwtToken
        },
        body: JSON.stringify({ name, description, quantity })
    })
    .then(response => response.json())
    .then(data => {
        if (data.result) {
            alert('Товар обновлён');
            loadItems();
        } else {
            alert('Ошибка: ' + data.error);
        }
    })
    .catch(err => console.error(err));
}

function deleteItem(id) {
    if (confirm('Удалить товар?')) {
        fetch(`/items/${id}`, {
            method: 'DELETE',
            headers: {
                'Authorization': 'Bearer ' + jwtToken
            }
        })
        .then(response => response.json())
        .then(data => {
            if (data.result) {
                alert('Товар удалён');
                loadItems();
            } else {
                alert('Ошибка: ' + data.error);
            }
        })
        .catch(err => console.error(err));
    }
}

function loadHistory(itemId) {
    fetch(`/items/${itemId}/history`, {
        headers: {
            'Authorization': 'Bearer ' + jwtToken
        }
    })
    .then(response => response.json())
    .then(data => {
        if (data.history) {
            const tbody = document.getElementById('history-body');
            tbody.innerHTML = '';

            data.history.forEach(entry => {
                const row = document.createElement('tr');
                row.innerHTML = `
                    <td>${entry.action}</td>
                    <td>${entry.old_values || 'N/A'}</td>
                    <td>${entry.new_values || 'N/A'}</td>
                    <td>${entry.changed_by}</td>
                    <td>${entry.changed_at}</td>
                `;
                tbody.appendChild(row);
            });

            document.getElementById('history-section').style.display = 'block';
        }
    })
    .catch(err => console.error(err));
}