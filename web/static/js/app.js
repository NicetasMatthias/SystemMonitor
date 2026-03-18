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
        updateCharts(data.System);
        updateNetworkStatus(data.NetworkStats);
        updateDiskStatus(data.DiskStats)
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
    const container = document.getElementById('networkStatus');
    
    container.innerHTML = '';
    
    const sortedTargets = Object.entries(stats).sort((a, b) => 
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

function updateDiskStatus(diskStats) {
    const container = document.getElementById('diskStatus');
    
    if (!container) return;
    
    container.innerHTML = '';
    
    for (const [diskPath, stats] of Object.entries(diskStats)) {
        const item = document.createElement('div');
        item.className = 'disk-item';
        
        // Рассчитываем процент использования
        const usedPercent = ((stats.Used / stats.Total) * 100).toFixed(1);
        
        // Определяем цвет в зависимости от заполненности
        let statusClass = 'disk-ok';
        if (usedPercent > 90) {
            statusClass = 'disk-critical';
        } else if (usedPercent > 75) {
            statusClass = 'disk-warning';
        }
        
        // Форматируем размеры в ГБ с одним знаком после запятой
        const usedGB = stats.Used.toFixed(1);
        const totalGB = stats.Total.toFixed(1);
        const freeGB = (stats.Total - stats.Used).toFixed(1);
        
        item.innerHTML = `
            <div class="disk-info">
                <div class="disk-name">${diskPath}</div>
                <div class="disk-stats">
                    Used: ${usedGB} GB / ${totalGB} GB (${usedPercent}%)
                </div>
                <div class="disk-bar">
                    <div class="disk-bar-fill ${statusClass}" style="width: ${usedPercent}%"></div>
                </div>
                <div class="disk-free">Free: ${freeGB} GB</div>
            </div>
        `;
        
        container.appendChild(item);
    }
}