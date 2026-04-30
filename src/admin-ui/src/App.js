import React, { useState, useEffect, useRef } from 'react';
import {
    CssBaseline,
    Container,
    Typography,
    Button,
    Box,
    Paper,
    LinearProgress,
    Grid,
    Card,
    CardContent,
    AppBar,
    Toolbar,
    Badge,
    Chip,
    Divider,
    Stack,
    Alert
} from '@mui/material';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import StopIcon from '@mui/icons-material/Stop';
import PauseIcon from '@mui/icons-material/Pause';
import SpeedIcon from '@mui/icons-material/Speed';
import StorageIcon from '@mui/icons-material/Storage';
import {
    LineChart,
    Line,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    Legend,
    ResponsiveContainer
} from 'recharts';

function App() {
    const [isRunning, setIsRunning] = useState(false);
    const [status, setStatus] = useState(null);
    const [loading, setLoading] = useState(false);
    const [chartData, setChartData] = useState([]);
    const [startTime, setStartTime] = useState(null);
    const startTimeRef = useRef(null);
    const wasRunningRef = useRef(false);
    const minuteStartTotalsRef = useRef({});
    const [channelPerformance, setChannelPerformance] = useState([
        { channel: 'email', successRate: 98.5, avgLatency: 245, messages: 42000 },
        { channel: 'sms', successRate: 99.2, avgLatency: 128, messages: 36000 },
        { channel: 'whatsapp', successRate: 97.1, avgLatency: 156, messages: 30000 },
        { channel: 'ota', successRate: 96.8, avgLatency: 189, messages: 12000 }
    ]);
    const API_BASE = process.env.REACT_APP_API_URL || 'http://localhost:8086';

    const fetchStatus = async () => {
        try {
            const response = await fetch(`${API_BASE}/simulator/status`);
            const data = await response.json();
            if (data.data) {
                setStatus(data.data);
                setIsRunning(data.data.IsRunning);

                const isRunning = data.data.IsRunning;
                const elapsedSeconds = data.data.ElapsedSeconds;
                const totalGenerated = data.data.TotalGenerated;

                // Update chart data
                if (isRunning) {
                    // Initialize tracking ref on first data point
                    if (!startTimeRef.current) {
                        startTimeRef.current = {
                            lastTotal: totalGenerated,
                            lastElapsed: elapsedSeconds,
                            lastPointTime: -1
                        };
                        console.log('🔄 Chart initialized at elapsed=', elapsedSeconds);
                        setChartData([]);
                    } else {
                        // Check if we have a new time point (not a duplicate)
                        if (elapsedSeconds > startTimeRef.current.lastPointTime) {
                            // Calculate messages and time delta since last update
                            const messageDelta = totalGenerated - startTimeRef.current.lastTotal;
                            const timeDelta = elapsedSeconds - startTimeRef.current.lastElapsed;

                            // Only add point if we have meaningful time delta
                            if (timeDelta > 0) {
                                // Update tracking for next iteration
                                startTimeRef.current.lastTotal = totalGenerated;
                                startTimeRef.current.lastElapsed = elapsedSeconds;
                                startTimeRef.current.lastPointTime = elapsedSeconds;

                                // Calculate rate (messages per minute)
                                const ratePerMinute = Math.round((messageDelta * 60) / timeDelta);

                                const newPoint = {
                                    time: elapsedSeconds,
                                    email: Math.round(ratePerMinute * 0.35),
                                    sms: Math.round(ratePerMinute * 0.30),
                                    whatsapp: Math.round(ratePerMinute * 0.25),
                                    ota: Math.round(ratePerMinute * 0.10)
                                };

                                console.log(`📊 Point at t=${elapsedSeconds}s: email=${newPoint.email}, sms=${newPoint.sms}, whatsapp=${newPoint.whatsapp}, ota=${newPoint.ota}`);

                                setChartData(prev => [...prev, newPoint]);
                            }
                        }
                    }
                } else {
                    // Test stopped - keep existing data frozen
                    if (wasRunningRef.current === true) {
                        console.log('✋ Test stopped at', elapsedSeconds, 's with', totalGenerated, 'messages');
                        wasRunningRef.current = false;
                    }
                }

                // Update channel performance
                if (isRunning && startTime) {
                    const elapsedSec = Math.floor((Date.now() - startTime) / 1000);
                    const progress = Math.min(elapsedSec / 300, 1);

                    setChannelPerformance(prevPerf =>
                        prevPerf.map(channel => ({
                            ...channel,
                            avgLatency: Math.round(channel.avgLatency + (Math.random() * 40 - 20)),
                            successRate: Math.max(95, channel.successRate - (progress * 2) + (Math.random() * 3 - 1.5))
                        }))
                    );
                }
            }
        } catch (error) {
            console.error('Failed to fetch status:', error);
        }
    };

    useEffect(() => {
        fetchStatus();
        const interval = setInterval(fetchStatus, 2000);
        return () => clearInterval(interval);
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    useEffect(() => {
        console.log('✅ chartData updated:', chartData.length, 'points');
    }, [chartData]);

    const handleStart = async () => {
        setLoading(true);
        setChartData([]);
        setStartTime(null);
        startTimeRef.current = null;
        wasRunningRef.current = true;
        minuteStartTotalsRef.current = {};
        // Reset channel performance to baseline
        setChannelPerformance([
            { channel: 'email', successRate: 98.5, avgLatency: 245, messages: 42000 },
            { channel: 'sms', successRate: 99.2, avgLatency: 128, messages: 36000 },
            { channel: 'whatsapp', successRate: 97.1, avgLatency: 156, messages: 30000 },
            { channel: 'ota', successRate: 96.8, avgLatency: 189, messages: 12000 }
        ]);
        try {
            const response = await fetch(`${API_BASE}/simulator/start?duration=120`);
            const data = await response.json();
            console.log('Started:', data);
            fetchStatus();
        } catch (error) {
            console.error('Failed to start simulator:', error);
        }
        setLoading(false);
    };

    const handleStop = async () => {
        setLoading(true);
        try {
            const response = await fetch(`${API_BASE}/simulator/stop`);
            const data = await response.json();
            console.log('Stopped:', data);
            setStartTime(null);
            startTimeRef.current = null;
            wasRunningRef.current = false;
            minuteStartTotalsRef.current = {};
            // Reset channel performance to baseline
            setChannelPerformance([
                { channel: 'email', successRate: 98.5, avgLatency: 245, messages: 42000 },
                { channel: 'sms', successRate: 99.2, avgLatency: 128, messages: 36000 },
                { channel: 'whatsapp', successRate: 97.1, avgLatency: 156, messages: 30000 },
                { channel: 'ota', successRate: 96.8, avgLatency: 189, messages: 12000 }
            ]);
            fetchStatus();
        } catch (error) {
            console.error('Failed to stop simulator:', error);
        }
        setLoading(false);
    };

    return (
        <>
            <CssBaseline />

            {/* Header AppBar */}
            <AppBar
                position="static"
                sx={{
                    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                    boxShadow: '0 4px 20px rgba(0,0,0,0.1)'
                }}
            >
                <Toolbar>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, flex: 1 }}>
                        <SpeedIcon sx={{ fontSize: 32 }} />
                        <Box>
                            <Typography variant="h5" sx={{ fontWeight: 700, letterSpacing: -0.5 }}>
                                Traffic Simulator
                            </Typography>
                            <Typography variant="caption" sx={{ opacity: 0.9 }}>
                                Real-time Load Testing Dashboard
                            </Typography>
                        </Box>
                    </Box>
                    <Badge
                        badgeContent={isRunning ? "LIVE" : "IDLE"}
                        color={isRunning ? "success" : "default"}
                        sx={{
                            '& .MuiBadge-badge': {
                                backgroundColor: isRunning ? '#4caf50' : '#999',
                                color: '#fff',
                                boxShadow: `0 0 0 2px #fff, 0 0 10px ${isRunning ? '#4caf50' : '#999'}`
                            }
                        }}
                    >
                        <Box sx={{ width: 12, height: 12, borderRadius: '50%' }} />
                    </Badge>
                </Toolbar>
            </AppBar>

            <Container maxWidth="lg" sx={{ py: 4 }}>
                {/* Control Panel Card */}
                <Card
                    sx={{
                        mb: 3,
                        background: 'linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%)',
                        border: '1px solid rgba(102, 126, 234, 0.2)'
                    }}
                >
                    <CardContent>
                        <Grid container spacing={2} alignItems="center">
                            <Grid item xs={12} md={8}>
                                <Typography variant="h6" sx={{ mb: 2, fontWeight: 600 }}>
                                    Test Controls
                                </Typography>
                                <Stack direction="row" spacing={1.5} sx={{ flexWrap: 'wrap', gap: 1 }}>
                                    <Button
                                        variant="contained"
                                        size="large"
                                        startIcon={<PlayArrowIcon />}
                                        onClick={handleStart}
                                        disabled={isRunning || loading}
                                        sx={{
                                            background: 'linear-gradient(135deg, #4caf50 0%, #45a049 100%)',
                                            fontWeight: 600,
                                            textTransform: 'uppercase',
                                            fontSize: '0.9rem',
                                            '&:hover': {
                                                background: 'linear-gradient(135deg, #45a049 0%, #3d8b40 100%)',
                                            }
                                        }}
                                    >
                                        Start
                                    </Button>
                                    <Button
                                        variant="contained"
                                        size="large"
                                        startIcon={<PauseIcon />}
                                        disabled={!isRunning}
                                        sx={{
                                            background: 'linear-gradient(135deg, #ff9800 0%, #e68900 100%)',
                                            fontWeight: 600,
                                            textTransform: 'uppercase',
                                            fontSize: '0.9rem',
                                        }}
                                    >
                                        Pause
                                    </Button>
                                    <Button
                                        variant="contained"
                                        size="large"
                                        startIcon={<StopIcon />}
                                        onClick={handleStop}
                                        disabled={!isRunning || loading}
                                        sx={{
                                            background: 'linear-gradient(135deg, #f44336 0%, #da190b 100%)',
                                            fontWeight: 600,
                                            textTransform: 'uppercase',
                                            fontSize: '0.9rem',
                                        }}
                                    >
                                        Stop
                                    </Button>
                                </Stack>
                                {loading && <LinearProgress sx={{ mt: 2 }} />}
                            </Grid>
                            <Grid item xs={12} md={4} sx={{ textAlign: { xs: 'left', md: 'right' } }}>
                                <Stack spacing={0.5}>
                                    <Typography variant="body2" color="textSecondary">
                                        Status
                                    </Typography>
                                    <Chip
                                        label={isRunning ? '▶ RUNNING' : '⏹ IDLE'}
                                        color={isRunning ? 'success' : 'default'}
                                        variant="outlined"
                                        sx={{ fontWeight: 700, fontSize: '0.9rem' }}
                                    />
                                </Stack>
                            </Grid>
                        </Grid>
                    </CardContent>
                </Card>

                {/* Charts Section */}
                <Grid container spacing={3} sx={{ mb: 3 }}>
                    {/* Channel Performance Chart - Full Width */}
                    <Grid item xs={12}>
                        <Card>
                            <CardContent>
                                <Typography variant="h6" sx={{ mb: 2, fontWeight: 600, display: 'flex', alignItems: 'center', gap: 1 }}>
                                    <SpeedIcon />
                                    Messages by Channel (0min - 2min)
                                </Typography>
                                <Divider sx={{ mb: 2 }} />
                                {chartData.length > 0 ? (
                                    <ResponsiveContainer width="100%" height={400}>
                                        <LineChart
                                            data={chartData}
                                            margin={{ top: 5, right: 30, left: 0, bottom: 5 }}
                                        >
                                            <CartesianGrid strokeDasharray="3 3" stroke="#e0e0e0" />
                                            <XAxis
                                                dataKey="time"
                                                type="number"
                                                domain={[0, 120]}
                                                label={{ value: 'Seconds', position: 'insideBottomRight', offset: -5 }}
                                                stroke="#666"
                                                style={{ fontSize: '12px' }}
                                            />
                                            <YAxis
                                                type="number"
                                                domain={[0, 5000]}
                                                label={{ value: 'Messages/min', angle: -90, position: 'insideLeft' }}
                                                stroke="#666"
                                            />
                                            <Tooltip
                                                formatter={(value) => value.toLocaleString()}
                                                contentStyle={{
                                                    backgroundColor: '#fff',
                                                    border: '1px solid #ccc',
                                                    borderRadius: '8px'
                                                }}
                                            />
                                            <Legend
                                                wrapperStyle={{ paddingTop: '20px' }}
                                                iconType="line"
                                            />
                                            <Line
                                                type="monotone"
                                                dataKey="email"
                                                stroke="#667eea"
                                                strokeWidth={2}
                                                dot={false}
                                                isAnimationActive={false}
                                            />
                                            <Line
                                                type="monotone"
                                                dataKey="sms"
                                                stroke="#764ba2"
                                                strokeWidth={2}
                                                dot={false}
                                                isAnimationActive={false}
                                            />
                                            <Line
                                                type="monotone"
                                                dataKey="whatsapp"
                                                stroke="#f093fb"
                                                strokeWidth={2}
                                                dot={false}
                                                isAnimationActive={false}
                                            />
                                            <Line
                                                type="monotone"
                                                dataKey="ota"
                                                stroke="#4facfe"
                                                strokeWidth={2}
                                                dot={false}
                                                isAnimationActive={false}
                                            />
                                        </LineChart>
                                    </ResponsiveContainer>
                                ) : (
                                    <Alert severity="info">
                                        Start a simulation to view real-time data
                                    </Alert>
                                )}
                            </CardContent>
                        </Card>
                    </Grid>
                </Grid>

                {/* Dashboard Links */}
                <Card sx={{ background: 'linear-gradient(135deg, rgba(79, 172, 254, 0.05) 0%, rgba(0, 242, 254, 0.05) 100%)' }}>
                    <CardContent>
                        <Typography variant="h6" sx={{ mb: 2, fontWeight: 600 }}>
                            Dashboards in Real Time
                        </Typography>
                        <Divider sx={{ mb: 2 }} />
                        <Grid container spacing={2}>
                            <Grid item xs={12} sm={6} md={4}>
                                <Button
                                    fullWidth
                                    variant="outlined"
                                    href="http://localhost:3000"
                                    target="_blank"
                                    sx={{ fontWeight: 600 }}
                                >
                                    Grafana (Metrics & Logs)
                                </Button>
                            </Grid>
                            <Grid item xs={12} sm={6} md={4}>
                                <Button
                                    fullWidth
                                    variant="outlined"
                                    href="http://localhost:9090"
                                    target="_blank"
                                    sx={{ fontWeight: 600 }}
                                >
                                    Prometheus (Raw Metrics)
                                </Button>
                            </Grid>
                            <Grid item xs={12} sm={6} md={4}>
                                <Button
                                    fullWidth
                                    variant="outlined"
                                    href="http://localhost:3100"
                                    target="_blank"
                                    sx={{ fontWeight: 600 }}
                                >
                                    Loki (Logs)
                                </Button>
                            </Grid>
                        </Grid>
                    </CardContent>
                </Card>
            </Container>
        </>
    );
}

export default App;
