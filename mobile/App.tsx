import { StatusBar } from 'expo-status-bar';
import { useEffect, useState } from 'react';
import { ActivityIndicator, StyleSheet, Text, View } from 'react-native';

const API_URL = process.env.EXPO_PUBLIC_API_URL ?? 'http://localhost:8080';

export default function App() {
  const [health, setHealth] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function checkHealth() {
      try {
        const res = await fetch(`${API_URL}/health`);
        if (!res.ok) {
          throw new Error(`HTTP ${res.status}`);
        }
        const data = (await res.json()) as { status: string };
        if (!cancelled) {
          setHealth(data.status);
          setError(null);
        }
      } catch {
        if (!cancelled) {
          setHealth(null);
          setError('Backend unreachable. Start: cd backend && go run main.go');
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    checkHealth();
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <View style={styles.container}>
      <Text style={styles.eyebrow}>Cursor Hackathon Istanbul</Text>
      <Text style={styles.title}>Urban AI — Mobile</Text>
      <Text style={styles.lead}>
        Expo app for the hackathon demo. No raw imagery or PII in the repo.
      </Text>

      <View style={styles.card}>
        <Text style={styles.cardTitle}>Backend health</Text>
        <Text style={styles.apiUrl}>API: {API_URL}</Text>
        {loading ? (
          <ActivityIndicator style={styles.spinner} />
        ) : health ? (
          <Text style={styles.ok}>Status: {health}</Text>
        ) : (
          <Text style={styles.error}>{error}</Text>
        )}
      </View>

      <StatusBar style="auto" />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
    padding: 24,
    justifyContent: 'center',
  },
  eyebrow: {
    fontSize: 12,
    fontWeight: '600',
    textTransform: 'uppercase',
    letterSpacing: 1,
    color: '#666',
    marginBottom: 8,
  },
  title: {
    fontSize: 28,
    fontWeight: '700',
    marginBottom: 8,
    color: '#111',
  },
  lead: {
    fontSize: 16,
    lineHeight: 22,
    color: '#444',
    marginBottom: 24,
  },
  card: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 20,
    borderWidth: 1,
    borderColor: '#e5e5e5',
  },
  cardTitle: {
    fontSize: 16,
    fontWeight: '600',
    marginBottom: 8,
  },
  apiUrl: {
    fontSize: 13,
    color: '#666',
    marginBottom: 12,
  },
  spinner: {
    marginTop: 4,
  },
  ok: {
    color: '#0a7a2f',
    fontWeight: '600',
  },
  error: {
    color: '#b00020',
    lineHeight: 20,
  },
});
