let cpuChart, memoryChart;
let updateInterval = 5000;

document.addEventListener('DOMContentLoaded', () => {
    fetchStats();
    setInterval(fetchStats, updateInterval);
});

async function fetchStats() {
    try {
        const response = await fetch('/api/stats');
        const data = await response.json();
        updateCharts(data);
        updateNetworkStatus(data);
    } catch (error) {
        console.error('Error fetching stats:', error);
    }
}

function updateCharts(stats) {
    if (stats.length === 0) return;
    
    const timestamps = stats.map(s => {
        const date = new Date(s.Timestamp);
        return date.toLocaleTimeString();
    });
    
    const cpuData = stats.map(s => s.CPUUsage);
    const memoryData = stats.map(s => 
        ((s.MemoryUsed / s.MemoryTotal) * 100).toFixed(2)
    );
    
    if (!cpuChart) {
        const ctx = document.getElementById('cpuChart').getContext('2d');
        cpuChart = new Chart(ctx, {
            type: 'line',
            data: {
                labels: timestamps,
                datasets: [{
                    label: 'CPU Usage %',
                    data: cpuData,
                    borderColor: '#3498db',
                    backgroundColor: 'rgba(52, 152, 219, 0.1)',
                    tension: 0.4,
                    fill: true
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true,
                        max: 100
                    }
                }
            }
        });
    } else {
        cpuChart.data.labels = timestamps;
        cpuChart.data.datasets[0].data = cpuData;
        cpuChart.update();
    }
    
    if (!memoryChart) {
        const ctx = document.getElementById('memoryChart').getContext('2d');
        memoryChart = new Chart(ctx, {
            type: 'line',
            data: {
                labels: timestamps,
                datasets: [{
                    label: 'Memory Usage %',
                    data: memoryData,
                    borderColor: '#e74c3c',
                    backgroundColor: 'rgba(231, 76, 60, 0.1)',
                    tension: 0.4,
                    fill: true
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true,
                        max: 100
                    }
                }
            }
        });
    } else {
        memoryChart.data.labels = timestamps;
        memoryChart.data.datasets[0].data = memoryData;
        memoryChart.update();
    }
}

function updateNetworkStatus(stats) {
    if (stats.length === 0) return;
    
    const latestStats = stats[stats.length - 1];
    const networks = latestStats.Networks || {};
    const container = document.getElementById('networkStatus');
    
    container.innerHTML = '';
    
    const sortedTargets = Object.entries(networks).sort((a, b) => 
        a[0].localeCompare(b[0])
    );
    
    for (const [name, status] of sortedTargets) {
        const item = document.createElement('div');
        item.className = `network-item ${status.Reachable ? 'reachable' : 'unreachable'}`;
        
        const lastCheck = status.LastCheck ? new Date(status.LastCheck) : new Date();
        const timeStr = lastCheck.toLocaleTimeString();
        
        item.innerHTML = `
            <div>
                <div class="name">${name}</div>
                <div class="time">Last check: ${timeStr}</div>
            </div>
            <div class="status">
                ${status.Reachable ? 'Online' : 'Offline'}
                ${status.Latency ? `<br>${(status.Latency / 1000000).toFixed(2)}ms` : ''}
            </div>
        `;
        
        container.appendChild(item);
    }
}