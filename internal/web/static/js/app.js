let cpuChart = null;
let loadChart = null;
let memoryChart = null;

const updateInterval = 5000;

let logEventSource = null;
let logs = [];

const maxDisplayedLogs = 1000;

document.addEventListener("DOMContentLoaded", () => {
    fetchStats();
    setInterval(fetchStats, updateInterval);

    initializeLogs();
});


/*
 * --------------------------------------------------------------------------
 * Statistics
 * --------------------------------------------------------------------------
 */

async function fetchStats() {
    try {
        const [
            cpuResponse,
            memoryResponse,
            diskResponse,
            networkResponse,
            systemResponse
        ] = await Promise.all([
            fetch("/api/stats/cpu"),
            fetch("/api/stats/memory"),
            fetch("/api/stats/disk"),
            fetch("/api/stats/network"),
            fetch("/api/stats/system")
        ]);

        if (
            !cpuResponse.ok ||
            !memoryResponse.ok ||
            !diskResponse.ok ||
            !networkResponse.ok ||
            !systemResponse.ok
        ) {
            throw new Error("Failed to fetch system statistics");
        }

        const [
            cpu,
            memory,
            disk,
            network,
            system
        ] = await Promise.all([
            cpuResponse.json(),
            memoryResponse.json(),
            diskResponse.json(),
            networkResponse.json(),
            systemResponse.json()
        ]);

        updateCPU(cpu);
        updateMemory(memory);
        updateDisk(disk);
        updateNetwork(network);
        updateSystem(system);

        setConnectionStatus(true);

        document.getElementById("lastUpdate").textContent =
            `Updated ${new Date().toLocaleTimeString()}`;

    } catch (error) {
        console.error("Error fetching stats:", error);

        setConnectionStatus(false);

        document.getElementById("lastUpdate").textContent =
            "Connection error";
    }
}


/*
 * --------------------------------------------------------------------------
 * Connection
 * --------------------------------------------------------------------------
 */

function setConnectionStatus(connected) {
    const indicator = document.getElementById("connectionStatus");

    if (!indicator) {
        return;
    }

    indicator.classList.toggle("connected", connected);
    indicator.classList.toggle("disconnected", !connected);
}


/*
 * --------------------------------------------------------------------------
 * CPU
 * --------------------------------------------------------------------------
 */

function updateCPU(stats) {
    if (!stats || !Array.isArray(stats.history)) {
        return;
    }

    const history = stats.history.filter(sample =>
        sample &&
        sample.data &&
        sample.timestamp
    );

    if (history.length === 0) {
        return;
    }

    updateCPUChart(history);
    updateLoadChart(history);
    updateCoreUsage(history[history.length - 1]);
}

function updateCPUChart(history) {
    const labels = history.map(sample =>
        formatTime(sample.timestamp)
    );

    const data = history.map(sample => {
        if (!sample.data.usage || !sample.data.usage.valid) {
            return null;
        }

        return sample.data.usage.data;
    });

    if (!cpuChart) {
        const context =
            document.getElementById("cpuChart").getContext("2d");

        cpuChart = new Chart(context, {
            type: "line",

            data: {
                labels,

                datasets: [{
                    label: "CPU Usage %",
                    data,

                    borderColor: "#3498db",
                    backgroundColor: "rgba(52, 152, 219, 0.12)",

                    borderWidth: 2,
                    pointRadius: 0,

                    tension: 0.35,
                    fill: true
                }]
            },

            options: {
                responsive: true,
                maintainAspectRatio: false,

                animation: false,

                scales: {
                    y: {
                        beginAtZero: true,
                        max: 100,

                        ticks: {
                            callback: value => `${value}%`
                        }
                    }
                }
            }
        });

        return;
    }

    cpuChart.data.labels = labels;
    cpuChart.data.datasets[0].data = data;
    cpuChart.update("none");
}

function updateLoadChart(history) {
    const labels = history.map(sample =>
        formatTime(sample.timestamp)
    );

    const load1 = history.map(sample => {
        if (!sample.data.load || !sample.data.load.valid) {
            return null;
        }

        return sample.data.load.data.load1;
    });

    const load5 = history.map(sample => {
        if (!sample.data.load || !sample.data.load.valid) {
            return null;
        }

        return sample.data.load.data.load5;
    });

    const load15 = history.map(sample => {
        if (!sample.data.load || !sample.data.load.valid) {
            return null;
        }

        return sample.data.load.data.load15;
    });

    const datasets = [
        {
            label: "1 min",
            data: load1,
            borderColor: "#e74c3c",
            backgroundColor: "transparent",
            borderWidth: 2,
            pointRadius: 0,
            tension: 0.35
        },
        {
            label: "5 min",
            data: load5,
            borderColor: "#f39c12",
            backgroundColor: "transparent",
            borderWidth: 2,
            pointRadius: 0,
            tension: 0.35
        },
        {
            label: "15 min",
            data: load15,
            borderColor: "#27ae60",
            backgroundColor: "transparent",
            borderWidth: 2,
            pointRadius: 0,
            tension: 0.35
        }
    ];

    if (!loadChart) {
        const context =
            document.getElementById("loadChart").getContext("2d");

        loadChart = new Chart(context, {
            type: "line",

            data: {
                labels,
                datasets
            },

            options: {
                responsive: true,
                maintainAspectRatio: false,

                animation: false,

                scales: {
                    y: {
                        beginAtZero: true
                    }
                }
            }
        });

        return;
    }

    loadChart.data.labels = labels;
    loadChart.data.datasets = datasets;
    loadChart.update("none");
}

function updateCoreUsage(sample) {
    const container = document.getElementById("coreUsage");
    const count = document.getElementById("coreCount");

    container.innerHTML = "";

    if (
        !sample ||
        !sample.data ||
        !sample.data.cores ||
        !sample.data.cores.valid
    ) {
        count.textContent = "Unavailable";
        return;
    }

    const cores = sample.data.cores.data || [];

    count.textContent = `${cores.length} cores`;

    for (const core of cores) {
        const usage = Number(core.usage) || 0;

        const item = document.createElement("div");
        item.className = "core-item";

        item.innerHTML = `
            <div class="core-header">
                <span>Core ${escapeHtml(core.id)}</span>
                <strong>${usage.toFixed(1)}%</strong>
            </div>

            <div class="progress-bar">
                <div
                    class="progress-fill ${getUsageClass(usage)}"
                    style="width: ${Math.min(usage, 100)}%"
                ></div>
            </div>
        `;

        container.appendChild(item);
    }
}


/*
 * --------------------------------------------------------------------------
 * Memory
 * --------------------------------------------------------------------------
 */

function updateMemory(stats) {
    if (!stats || !Array.isArray(stats.history)) {
        return;
    }

    const history = stats.history;

    if (history.length === 0) {
        return;
    }

    const latest = history[history.length - 1];

    if (!latest.valid) {
        return;
    }

    document.getElementById("memoryUsed").textContent =
        formatBytes(latest.used);

    document.getElementById("memoryAvailable").textContent =
        formatBytes(latest.available);

    document.getElementById("memoryTotal").textContent =
        formatBytes(latest.total);

    document.getElementById("memoryPercentage").textContent =
        `${Number(latest.percentage).toFixed(1)}%`;

    updateMemoryChart(history);
}

function updateMemoryChart(history) {
    const labels = history.map(sample =>
        formatTime(sample.timestamp)
    );

    const data = history.map(sample => {
        if (!sample.valid) {
            return null;
        }

        return sample.percentage;
    });

    if (!memoryChart) {
        const context =
            document.getElementById("memoryChart").getContext("2d");

        memoryChart = new Chart(context, {
            type: "line",

            data: {
                labels,

                datasets: [{
                    label: "Memory Usage %",
                    data,

                    borderColor: "#9b59b6",
                    backgroundColor: "rgba(155, 89, 182, 0.12)",

                    borderWidth: 2,
                    pointRadius: 0,

                    tension: 0.35,
                    fill: true
                }]
            },

            options: {
                responsive: true,
                maintainAspectRatio: false,

                animation: false,

                scales: {
                    y: {
                        beginAtZero: true,
                        max: 100,

                        ticks: {
                            callback: value => `${value}%`
                        }
                    }
                }
            }
        });

        return;
    }

    memoryChart.data.labels = labels;
    memoryChart.data.datasets[0].data = data;
    memoryChart.update("none");
}


/*
 * --------------------------------------------------------------------------
 * Disk
 * --------------------------------------------------------------------------
 */

function updateDisk(stats) {
    if (!stats) {
        return;
    }

    updateMountPoints(stats.mount_points || []);
    updateDiskDevices(stats.devices || {});
}

function updateMountPoints(mountPoints) {
    const container = document.getElementById("diskStatus");
    const count = document.getElementById("mountPointCount");

    container.innerHTML = "";
    count.textContent = mountPoints.length;

    if (mountPoints.length === 0) {
        container.innerHTML =
            '<div class="empty-state">No mount points available</div>';

        return;
    }

    for (const mount of mountPoints) {
        const capacity = mount.capacity || {};

        const usedPercent = capacity.valid
            ? Number(capacity.used_percent)
            : 0;

        const item = document.createElement("div");
        item.className = "disk-item";

        item.innerHTML = `
            <div class="disk-header">
                <div>
                    <div class="disk-name">
                        ${escapeHtml(mount.mountpoint || "Unknown")}
                    </div>

                    <div class="disk-device">
                        ${escapeHtml(mount.device || "Unknown")}
                    </div>
                </div>

                <span class="filesystem">
                    ${escapeHtml(mount.fs || "Unknown")}
                </span>
            </div>

            ${capacity.valid
                ? `
                    <div class="disk-stats">
                        <span>
                            ${formatBytes(capacity.used_bytes)}
                            /
                            ${formatBytes(capacity.total_bytes)}
                        </span>

                        <strong>
                            ${usedPercent.toFixed(1)}%
                        </strong>
                    </div>

                    <div class="progress-bar disk-progress">
                        <div
                            class="progress-fill ${getUsageClass(usedPercent)}"
                            style="width: ${Math.min(usedPercent, 100)}%"
                        ></div>
                    </div>
                `
                : `
                    <div class="unavailable">
                        Capacity information unavailable
                    </div>
                `
            }
        `;

        container.appendChild(item);
    }
}

function updateDiskDevices(devices) {
    const container = document.getElementById("diskDevices");

    container.innerHTML = "";

    const entries = Object.entries(devices);

    if (entries.length === 0) {
        container.innerHTML =
            '<div class="empty-state">No disk I/O data available</div>';

        return;
    }

    for (const [deviceName, device] of entries) {
        const history = device.history || [];

        const latest =
            history.length > 0
                ? history[history.length - 1]
                : null;

        const item = document.createElement("div");
        item.className = "device-item";

        if (!latest || !latest.valid) {
            item.innerHTML = `
                <div class="device-name">
                    ${escapeHtml(deviceName)}
                </div>

                <div class="unavailable">
                    I/O data unavailable
                </div>
            `;

            container.appendChild(item);
            continue;
        }

        item.innerHTML = `
            <div class="device-header">
                <span class="device-name">
                    ${escapeHtml(deviceName)}
                </span>

                <span class="device-time">
                    ${formatTime(latest.timestamp)}
                </span>
            </div>

            <div class="io-grid">
                <div class="io-stat">
                    <span class="io-label">Read</span>
                    <strong>
                        ${formatBytesPerSecond(latest.read_throughput)}
                    </strong>
                </div>

                <div class="io-stat">
                    <span class="io-label">Write</span>
                    <strong>
                        ${formatBytesPerSecond(latest.write_throughput)}
                    </strong>
                </div>

                <div class="io-stat">
                    <span class="io-label">Read IOPS</span>
                    <strong>
                        ${formatNumber(latest.read_iops)}
                    </strong>
                </div>

                <div class="io-stat">
                    <span class="io-label">Write IOPS</span>
                    <strong>
                        ${formatNumber(latest.write_iops)}
                    </strong>
                </div>
            </div>
        `;

        container.appendChild(item);
    }
}


/*
 * --------------------------------------------------------------------------
 * Network
 * --------------------------------------------------------------------------
 */

function updateNetwork(stats) {
    const container = document.getElementById("networkStatus");
    const count = document.getElementById("networkTargetCount");

    container.innerHTML = "";

    const targets = stats && stats.stats
        ? Object.entries(stats.stats)
        : [];

    count.textContent = targets.length;

    if (targets.length === 0) {
        container.innerHTML =
            '<div class="empty-state">No network targets configured</div>';

        return;
    }

    targets.sort((a, b) =>
        a[0].localeCompare(b[0])
    );

    for (const [name, status] of targets) {
        const reachable = Boolean(status.reachable);

        const item = document.createElement("div");

        item.className =
            `network-item ${reachable ? "reachable" : "unreachable"}`;

        const latency =
            reachable && status.latency
                ? formatDuration(status.latency)
                : "—";

        item.innerHTML = `
            <div class="network-main">
                <div class="network-name">
                    ${escapeHtml(name)}
                </div>

                <div class="network-time">
                    Last check:
                    ${status.last_check
                ? formatDateTime(status.last_check)
                : "—"}
                </div>
            </div>

            <div class="network-result">
                <span class="network-state">
                    ${reachable ? "Online" : "Offline"}
                </span>

                <span class="network-latency">
                    ${latency}
                </span>
            </div>
        `;

        container.appendChild(item);
    }
}


/*
 * --------------------------------------------------------------------------
 * System
 * --------------------------------------------------------------------------
 */

function updateSystem(stats) {
    if (!stats) {
        return;
    }

    const host = stats.host || {};
    const activity = stats.activity || {};

    setText("hostname", host.hostname);
    setText("os", host.os);

    setText(
        "platform",
        [host.platform, host.platform_version]
            .filter(Boolean)
            .join(" ")
    );

    setText("architecture", host.architecture);
    setText("kernelVersion", host.kernel_version);

    const cpuParts = [];

    if (host.logical_cp_us !== undefined) {
        cpuParts.push(`${host.logical_cp_us} logical`);
    }

    if (host.physical_cores !== undefined) {
        cpuParts.push(`${host.physical_cores} physical`);
    }

    setText(
        "cpuCount",
        cpuParts.length > 0
            ? cpuParts.join(" / ")
            : "—"
    );

    if (host.boot_time) {
        setText(
            "bootTime",
            formatDateTime(host.boot_time)
        );

        setText(
            "uptime",
            formatUptime(host.boot_time)
        );
    } else {
        setText("bootTime", "—");
        setText("uptime", "—");
    }

    setText(
        "processCount",
        activity.process_count !== undefined
            ? activity.process_count.toLocaleString()
            : "—"
    );

    setText(
        "sessionCount",
        activity.session_count !== undefined
            ? activity.session_count.toLocaleString()
            : "—"
    );
}


/*
 * --------------------------------------------------------------------------
 * Logs
 * --------------------------------------------------------------------------
 */

function initializeLogs() {
    const filter = document.getElementById("logLevelFilter");
    const autoScroll = document.getElementById("logAutoScroll");
    const clearButton = document.getElementById("clearLogs");

    filter.addEventListener("change", renderLogs);

    clearButton.addEventListener("click", () => {
        logs = [];
        renderLogs();
    });

    autoScroll.addEventListener("change", () => {
        if (autoScroll.checked) {
            scrollLogsToBottom();
        }
    });

    connectLogs();
}

function connectLogs() {
    if (logEventSource) {
        logEventSource.close();
    }

    setLogConnectionStatus(false, "Connecting...");

    logEventSource = new EventSource("/api/logs/stream");

    logEventSource.onopen = () => {
        setLogConnectionStatus(true, "Connected");
    };

    logEventSource.onmessage = event => {
        try {
            const entry = JSON.parse(event.data);

            addLog(entry);

            setLogConnectionStatus(true, "Connected");
        } catch (error) {
            console.error("Failed to parse log event:", error);
        }
    };

    logEventSource.onerror = () => {
        setLogConnectionStatus(false, "Reconnecting...");
    };
}

function setLogConnectionStatus(connected, text) {
    const indicator =
        document.getElementById("logConnectionStatus");

    const label =
        document.getElementById("logConnectionText");

    if (!indicator || !label) {
        return;
    }

    indicator.classList.toggle("connected", connected);
    indicator.classList.toggle("disconnected", !connected);

    label.textContent = text;
}

function addLog(entry) {
    if (!entry) {
        return;
    }

    logs.push(entry);

    if (logs.length > maxDisplayedLogs) {
        logs.splice(0, logs.length - maxDisplayedLogs);
    }

    renderLogs();
}

function renderLogs() {
    const container = document.getElementById("logs");
    const empty = document.getElementById("logsEmpty");
    const filter = document.getElementById("logLevelFilter");

    if (!container) {
        return;
    }

    const selectedLevel = filter
        ? filter.value
        : "all";

    const filteredLogs = selectedLevel === "all"
        ? logs
        : logs.filter(entry =>
            normalizeLogLevel(entry.level) === selectedLevel
        );

    container.innerHTML = "";

    if (filteredLogs.length === 0) {
        const emptyElement = document.createElement("div");

        emptyElement.className = "empty-state";
        emptyElement.textContent =
            logs.length === 0
                ? "Waiting for logs..."
                : "No logs match the selected level";

        container.appendChild(emptyElement);

        return;
    }

    for (const entry of filteredLogs) {
        container.appendChild(createLogElement(entry));
    }

    if (isAutoScrollEnabled()) {
        scrollLogsToBottom();
    }
}

function createLogElement(entry) {
    const item = document.createElement("div");

    const level = normalizeLogLevel(entry.level);

    item.className = `log-entry log-${level.toLowerCase()}`;

    const time = document.createElement("span");

    time.className = "log-time";
    time.textContent = formatDateTime(entry.time);

    const levelElement = document.createElement("span");

    levelElement.className = "log-level";
    levelElement.textContent = level;

    const message = document.createElement("span");

    message.className = "log-message";
    message.textContent = entry.message || "";

    item.appendChild(time);
    item.appendChild(levelElement);
    item.appendChild(message);

    if (
        entry.attrs &&
        typeof entry.attrs === "object" &&
        Object.keys(entry.attrs).length > 0
    ) {
        const attrs = document.createElement("span");

        attrs.className = "log-attrs";
        attrs.textContent = formatLogAttrs(entry.attrs);

        item.appendChild(attrs);
    }

    return item;
}

function formatLogAttrs(attrs) {
    return Object.entries(attrs)
        .map(([key, value]) => {
            return `${key}=${formatLogAttrValue(value)}`;
        })
        .join(" ");
}

function formatLogAttrValue(value) {
    if (value === null || value === undefined) {
        return "null";
    }

    if (typeof value === "object") {
        try {
            return JSON.stringify(value);
        } catch {
            return String(value);
        }
    }

    return String(value);
}

function normalizeLogLevel(level) {
    if (!level) {
        return "INFO";
    }

    const value = String(level).toUpperCase();

    if (value === "WARNING") {
        return "WARN";
    }

    if (value === "DEBUG") {
        return "DEBUG";
    }

    if (value === "WARN") {
        return "WARN";
    }

    if (value === "ERROR") {
        return "ERROR";
    }

    return "INFO";
}

function isAutoScrollEnabled() {
    const checkbox =
        document.getElementById("logAutoScroll");

    return !checkbox || checkbox.checked;
}

function scrollLogsToBottom() {
    const container =
        document.getElementById("logs");

    if (!container) {
        return;
    }

    requestAnimationFrame(() => {
        container.scrollTop = container.scrollHeight;
    });
}


/*
 * --------------------------------------------------------------------------
 * Formatting
 * --------------------------------------------------------------------------
 */

function formatBytes(bytes) {
    if (
        bytes === null ||
        bytes === undefined ||
        !Number.isFinite(Number(bytes))
    ) {
        return "—";
    }

    const value = Number(bytes);

    if (value === 0) {
        return "0 B";
    }

    const units = [
        "B",
        "KB",
        "MB",
        "GB",
        "TB",
        "PB"
    ];

    const exponent =
        Math.min(
            Math.floor(Math.log(value) / Math.log(1024)),
            units.length - 1
        );

    const converted =
        value / Math.pow(1024, exponent);

    return `${converted.toFixed(
        exponent === 0 ? 0 : 2
    )} ${units[exponent]}`;
}

function formatBytesPerSecond(bytes) {
    return `${formatBytes(bytes)}/s`;
}

function formatNumber(value) {
    if (
        value === null ||
        value === undefined ||
        !Number.isFinite(Number(value))
    ) {
        return "—";
    }

    return Number(value).toFixed(2);
}

function formatTime(timestamp) {
    if (!timestamp) {
        return "—";
    }

    return new Date(timestamp).toLocaleTimeString();
}

function formatDateTime(timestamp) {
    if (!timestamp) {
        return "—";
    }

    return new Date(timestamp).toLocaleString();
}

function formatDuration(nanoseconds) {
    if (!nanoseconds) {
        return "—";
    }

    const milliseconds = Number(nanoseconds) / 1e6;

    if (milliseconds < 1) {
        return `${(milliseconds * 1000).toFixed(0)} μs`;
    }

    if (milliseconds < 1000) {
        return `${milliseconds.toFixed(2)} ms`;
    }

    return `${(milliseconds / 1000).toFixed(2)} s`;
}

function formatUptime(bootTime) {
    const boot = new Date(bootTime);
    const now = new Date();

    let seconds =
        Math.floor(
            (now.getTime() - boot.getTime()) / 1000
        );

    if (seconds < 0) {
        return "—";
    }

    const days = Math.floor(seconds / 86400);
    seconds %= 86400;

    const hours = Math.floor(seconds / 3600);
    seconds %= 3600;

    const minutes = Math.floor(seconds / 60);
    seconds %= 60;

    const parts = [];

    if (days > 0) {
        parts.push(`${days}d`);
    }

    if (hours > 0 || days > 0) {
        parts.push(`${hours}h`);
    }

    if (minutes > 0 || hours > 0 || days > 0) {
        parts.push(`${minutes}m`);
    }

    parts.push(`${seconds}s`);

    return parts.join(" ");
}

function getUsageClass(value) {
    if (value >= 90) {
        return "critical";
    }

    if (value >= 75) {
        return "warning";
    }

    return "normal";
}


/*
 * --------------------------------------------------------------------------
 * DOM helpers
 * --------------------------------------------------------------------------
 */

function setText(id, value) {
    const element = document.getElementById(id);

    if (!element) {
        return;
    }

    element.textContent =
        value === undefined ||
            value === null ||
            value === ""
            ? "—"
            : value;
}

function escapeHtml(value) {
    if (value === undefined || value === null) {
        return "";
    }

    return String(value)
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;")
        .replaceAll('"', "&quot;")
        .replaceAll("'", "&#039;");
}
