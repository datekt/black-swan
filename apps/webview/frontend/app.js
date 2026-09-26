const canvas = document.getElementById('orbitCanvas');
const ctx = canvas.getContext('2d');
const statusEl = document.getElementById('status');
const speedInput = document.getElementById('speed');
const speedLabel = document.getElementById('speedLabel');
const togglePlayBtn = document.getElementById('togglePlay');
const restartBtn = document.getElementById('restart');
const frameInfoEl = document.getElementById('frameInfo');
const simTimeEl = document.getElementById('simTime');
const eventBanner = document.getElementById('eventBanner');

let simulationData = null;
let currentFrame = 0;
let animationId = null;
let playing = true;
let speed = 1;
let lastFrameTime = 0;
let accumulated = 0;
let bannerTimeout = null;

const BASE_FRAME_INTERVAL_MS = 1000 / 30;

function resizeCanvas() {
    const dpr = window.devicePixelRatio || 1;
    canvas.width = window.innerWidth * dpr;
    canvas.height = window.innerHeight * dpr;
    canvas.style.width = window.innerWidth + 'px';
    canvas.style.height = window.innerHeight + 'px';
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    draw();
}

function computeBounds(frame) {
    let minX = Infinity, maxX = -Infinity, minY = Infinity, maxY = -Infinity;
    for (const body of frame.bodies) {
        const x = body.position[0];
        const y = body.position[1];
        if (x < minX) minX = x;
        if (x > maxX) maxX = x;
        if (y < minY) minY = y;
        if (y > maxY) maxY = y;
    }
    if (!isFinite(minX)) {
        return { minX: -1, maxX: 1, minY: -1, maxY: 1 };
    }
    return { minX, maxX, minY, maxY };
}

function formatTime(seconds) {
    const days = seconds / 86400;
    if (days >= 1) {
        return days.toFixed(2) + ' days';
    }
    const hours = seconds / 3600;
    if (hours >= 1) {
        return hours.toFixed(2) + ' h';
    }
    return seconds.toFixed(0) + ' s';
}

function draw() {
    ctx.fillStyle = '#0d1117';
    ctx.fillRect(0, 0, window.innerWidth, window.innerHeight);

    if (!simulationData || simulationData.length === 0) {
        ctx.fillStyle = '#ffffff';
        ctx.font = '14px monospace';
        ctx.textAlign = 'center';
        ctx.fillText('[ BSE Engine Offline Canvas ready. Awaiting telemetry from Go core... ]', window.innerWidth / 2, window.innerHeight / 2);
        return;
    }

    const frame = simulationData[currentFrame];
    const bounds = computeBounds(frame);
    const width = Math.max(bounds.maxX - bounds.minX, 1);
    const height = Math.max(bounds.maxY - bounds.minY, 1);
    const padding = 0.12;
    const w = window.innerWidth * (1 - 2 * padding);
    const h = window.innerHeight * (1 - 2 * padding);
    const scale = Math.min(w / width, h / height);
    const cx = (bounds.minX + bounds.maxX) / 2;
    const cy = (bounds.minY + bounds.maxY) / 2;
    const offsetX = window.innerWidth / 2 - cx * scale;
    const offsetY = window.innerHeight / 2 + cy * scale;

    for (const body of frame.bodies) {
        const sx = body.position[0] * scale + offsetX;
        const sy = -body.position[1] * scale + offsetY;
        const scaledRadius = (body.radius || 1e6) * scale;
        const r = Math.max(2.5, Math.min(scaledRadius, 40));

        if (body.name && body.name.toLowerCase().includes('black hole')) {
            const grad = ctx.createRadialGradient(sx, sy, r * 0.4, sx, sy, r * 2.5);
            grad.addColorStop(0, 'rgba(138, 43, 226, 0.9)');
            grad.addColorStop(1, 'rgba(138, 43, 226, 0)');
            ctx.beginPath();
            ctx.arc(sx, sy, r * 2.5, 0, Math.PI * 2);
            ctx.fillStyle = grad;
            ctx.fill();
        }

        ctx.beginPath();
        ctx.arc(sx, sy, r, 0, Math.PI * 2);

        if (body.name === 'Sun') {
            ctx.fillStyle = '#ffcc00';
        } else if (body.name && body.name.toLowerCase().includes('black hole')) {
            ctx.fillStyle = '#0a0a12';
            ctx.strokeStyle = '#8a2be2';
            ctx.lineWidth = 2;
            ctx.fill();
            ctx.stroke();
        } else {
            ctx.fillStyle = '#58a6ff';
        }

        if (body.name !== 'Sun' && !(body.name && body.name.toLowerCase().includes('black hole'))) {
            ctx.fill();
        }

        ctx.fillStyle = '#c9d1d9';
        ctx.font = '11px monospace';
        ctx.textAlign = 'left';
        ctx.fillText(body.name, sx + r + 4, sy - r);
    }

    if (frameInfoEl) {
        frameInfoEl.textContent = 'Frame: ' + (currentFrame + 1) + ' / ' + simulationData.length + '  |  Step: ' + frame.step;
    }
    if (simTimeEl) {
        simTimeEl.textContent = 'Sim time: ' + formatTime(frame.sim_time_seconds);
    }
}

function showEvents(events) {
    if (!events || events.length === 0) {
        return;
    }
    const text = events.map(e => e.absorbed + ' absorbed by ' + e.absorber).join('  •  ');
    eventBanner.textContent = '💥 ' + text;
    eventBanner.classList.add('visible');
    if (bannerTimeout) {
        clearTimeout(bannerTimeout);
    }
    bannerTimeout = setTimeout(() => {
        eventBanner.classList.remove('visible');
    }, 2500);
}

function tick(timestamp) {
    if (!simulationData || simulationData.length === 0) {
        return;
    }

    if (lastFrameTime === 0) {
        lastFrameTime = timestamp;
    }

    const deltaMs = timestamp - lastFrameTime;
    lastFrameTime = timestamp;

    if (playing && speed > 0) {
        accumulated += deltaMs * speed;

        while (accumulated >= BASE_FRAME_INTERVAL_MS) {
            accumulated -= BASE_FRAME_INTERVAL_MS;
            currentFrame = (currentFrame + 1) % simulationData.length;
            const frame = simulationData[currentFrame];
            if (frame.events && frame.events.length > 0) {
                showEvents(frame.events);
            }
        }

        draw();
    }

    animationId = requestAnimationFrame(tick);
}

async function loadSimulation() {
    try {
        const response = await fetch('/api/trajectory');
        if (!response.ok) {
            throw new Error('trajectory not available');
        }
        simulationData = await response.json();
        if (!Array.isArray(simulationData) || simulationData.length === 0) {
            throw new Error('empty trajectory');
        }
        if (statusEl) {
            statusEl.textContent = 'Streaming ' + simulationData.length + ' frames';
            statusEl.style.color = '#3fb950';
        }
        if (animationId) {
            cancelAnimationFrame(animationId);
        }
        currentFrame = 0;
        accumulated = 0;
        lastFrameTime = 0;
        animationId = requestAnimationFrame(tick);
    } catch (err) {
        if (statusEl) {
            statusEl.textContent = 'No trajectory. Run: go run ./apps/cli';
            statusEl.style.color = '#d29922';
        }
        simulationData = null;
        draw();
    }
}

speedInput.addEventListener('input', () => {
    speed = parseFloat(speedInput.value);
    speedLabel.textContent = '×' + speed.toFixed(2);
    if (speed === 0) {
        playing = false;
        togglePlayBtn.textContent = 'Play';
    } else if (!playing) {
        playing = true;
        togglePlayBtn.textContent = 'Pause';
    }
});

togglePlayBtn.addEventListener('click', () => {
    playing = !playing;
    togglePlayBtn.textContent = playing ? 'Pause' : 'Play';
    lastFrameTime = 0;
});

restartBtn.addEventListener('click', () => {
    currentFrame = 0;
    accumulated = 0;
    lastFrameTime = 0;
    draw();
});

window.addEventListener('resize', resizeCanvas);
resizeCanvas();
loadSimulation();