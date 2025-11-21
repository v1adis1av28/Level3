
const API = "http://localhost:8080";

// ========= CREATE ITEM =========
document.getElementById("createForm").addEventListener("submit", async (e) => {
    e.preventDefault();

    const item = {
        name: document.getElementById("name").value,
        type: document.getElementById("type").value,
        price: Number(document.getElementById("price").value),
        date: document.getElementById("date").value || null,
    };

    const res = await fetch(`${API}/items`, {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify(item),
    });

    const data = await res.json();
    document.getElementById("createResult").innerText = JSON.stringify(data);
    loadItems();
});

// ========= LOAD ITEMS =========
async function loadItems() {
    const res = await fetch(`${API}/items`);
    const data = await res.json();
    const table = document.getElementById("itemsTable");
    table.innerHTML = "";

    data.items.forEach(item => {
        const row = `
            <tr>
                <td>${item.name}</td>
                <td>${item.price}</td>
                <td>${item.type}</td>
                <td>${new Date(item.date).toLocaleDateString()}</td>
            </tr>
        `;
        table.innerHTML += row;
    });
}

loadItems();

// ========= ANALYTICS =========
document.getElementById("analyticsForm").addEventListener("submit", async (e) => {
    e.preventDefault();

    let params = new URLSearchParams();
    params.append("type", document.getElementById("aType").value);

    const date = document.getElementById("aDate").value;
    const from = document.getElementById("aFrom").value;
    const to = document.getElementById("aTo").value;

    if (date) params.append("date", date);
    if (from) params.append("from", from);
    if (to) params.append("to", to);

    const res = await fetch(`${API}/analytics?${params.toString()}`);
    const data = await res.json();

    document.getElementById("analyticsResult").innerHTML = `
        <p><b>Тип:</b> ${data.analytics.type}</p>
        <p><b>Сумма:</b> ${data.analytics.sum}</p>
        <p><b>Среднее:</b> ${data.analytics.avg}</p>
        <p><b>Количество:</b> ${data.analytics.count}</p>
        <p><b>Медиана:</b> ${data.analytics.median}</p>
        <p><b>90 перцентиль:</b> ${data.analytics.percentile_90}</p>
    `;
});
