package main

const dashboardHTML = `
<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Mechanic Dashboard</title>
    <style>
        :root {
            --bg: #0d0d0d;
            --panel: #1a1a1a;
            --accent: #ff4d00;
            --text: #e0e0e0;
            --secondary: #00ccff;
            --success: #00ff88;
        }

        body {
            background-color: var(--bg);
            color: var(--text);
            font-family: 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            display: flex;
            flex-direction: column;
            align-items: center;
            min-height: 100vh;
        }

        .header {
            width: 100%;
            padding: 2rem;
            text-align: center;
            background: linear-gradient(180deg, #1f1f1f 0%, transparent 100%);
            border-bottom: 2px solid var(--accent);
        }

        .header h1 {
            margin: 0;
            letter-spacing: 4px;
            color: var(--accent);
            text-shadow: 0 0 10px rgba(255, 77, 0, 0.5);
        }

        .container {
            width: 90%;
            max-width: 1000px;
            margin-top: 2rem;
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 2rem;
        }

        @media (max-width: 768px) {
            .container { grid-template-columns: 1fr; }
        }

        .card {
            background: var(--panel);
            padding: 2rem;
            border-radius: 8px;
            border-left: 4px solid var(--accent);
            box-shadow: 0 10px 30px rgba(0,0,0,0.5);
        }

        .form-group {
            margin-bottom: 1.5rem;
        }

        label {
            display: block;
            margin-bottom: 0.5rem;
            color: var(--secondary);
            font-size: 0.9rem;
            text-transform: uppercase;
        }

        input {
            width: 100%;
            padding: 0.8rem;
            background: #2a2a2a;
            border: 1px solid #333;
            color: #fff;
            border-radius: 4px;
            outline: none;
            box-sizing: border-box;
        }

        input:focus {
            border-color: var(--accent);
        }

        button {
            width: 100%;
            padding: 1rem;
            background: var(--accent);
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-weight: bold;
            text-transform: uppercase;
            transition: 0.3s;
        }

        button:hover {
            background: #e64500;
            box-shadow: 0 0 15px rgba(255, 77, 0, 0.4);
        }

        .results {
            display: flex;
            flex-direction: column;
            gap: 1rem;
        }

        .stat-item {
            display: flex;
            justify-content: space-between;
            padding: 0.5rem 0;
            border-bottom: 1px solid #333;
        }

        .stat-value {
            font-weight: bold;
            color: var(--success);
        }

        .logs {
            grid-column: 1 / -1;
            background: #000;
            padding: 1rem;
            border-radius: 4px;
            font-family: 'Courier New', monospace;
            height: 200px;
            overflow-y: auto;
            border: 1px solid #333;
            font-size: 0.8rem;
        }

        .log-entry { margin-bottom: 0.2rem; }
        .log-time { color: #666; }
        .log-msg { color: #bbb; }
    </style>
</head>
<body>
    <div class="header">
        <h1>MECHANIC INC.</h1>
        <p>Advanced NetProbing Suite</p>
    </div>

    <div class="container">
        <div class="card">
            <h2>Configuration</h2>
            <div class="form-group">
                <label>Target URL</label>
                <input type="text" id="url" placeholder="https://example.com" value="http://example.com">
            </div>
            <div class="form-group">
                <label>Concurrency (Workers)</label>
                <input type="number" id="concurrency" value="10">
            </div>
            <div class="form-group">
                <label>Total Requests</label>
                <input type="number" id="count" value="100">
            </div>
            <button id="runBtn">Lancer le Probe</button>
        </div>

        <div class="card">
            <h2>Dernier Résultat</h2>
            <div id="resultContent" class="results">
                <p style="color: #666;">Aucun test effectué</p>
            </div>
        </div>

        <div class="logs" id="logs">
            <div class="log-entry"><span class="log-time">[H:M:S]</span> <span class="log-msg">Système prêt...</span></div>
        </div>
    </div>

    <script>
        const runBtn = document.getElementById('runBtn');
        const logs = document.getElementById('logs');
        const resultContent = document.getElementById('resultContent');

        function addLog(msg) {
            const time = new Date().toLocaleTimeString();
            const div = document.createElement('div');
            div.className = 'log-entry';
            div.innerHTML = ` + "`" + `<span class="log-time">[${time}]</span> <span class="log-msg">${msg}</span>` + "`" + `;
            logs.appendChild(div);
            logs.scrollTop = logs.scrollHeight;
        }

        runBtn.onclick = async () => {
            const url = document.getElementById('url').value;
            const concurrency = parseInt(document.getElementById('concurrency').value);
            const count = parseInt(document.getElementById('count').value);

            addLog(` + "`" + `Lancement du probe sur ${url}...` + "`" + `);
            runBtn.disabled = true;
            runBtn.innerText = "Traitement...";

            try {
                const res = await fetch('/api/probe', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({url, concurrency, count})
                });
                const data = await res.json();

                if (data.status === 'success') {
                    addLog(` + "`" + `Terminé ! Latence moy: ${data.latency_ms}ms` + "`" + `);
                    renderResults(data.results, data.latency_ms);
                }
            } catch (err) {
                addLog(` + "`" + `Erreur: ${err.message}` + "`" + `);
            } finally {
                runBtn.disabled = false;
                runBtn.innerText = "Lancer le Probe";
            }
        };

        function renderResults(res, lat) {
            resultContent.innerHTML = ` + "`" + `
                <div class="stat-item"><span>Total</span><span class="stat-value">${res.total_requests}</span></div>
                <div class="stat-item"><span>Succès</span><span class="stat-value" style="color: #00ff88">${res.success_count}</span></div>
                <div class="stat-item"><span>Échecs</span><span class="stat-value" style="color: #ff4d00">${res.error_count}</span></div>
                <div class="stat-item"><span>Latence Moy.</span><span class="stat-value">${lat}ms</span></div>
            ` + "`" + `;
        }
    </script>
</body>
</html>
`
