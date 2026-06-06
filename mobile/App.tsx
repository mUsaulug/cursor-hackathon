import { StatusBar } from 'expo-status-bar';
import { useCallback, useEffect, useState } from 'react';
import {
  ActivityIndicator,
  RefreshControl,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from 'react-native';

const API_URL = process.env.EXPO_PUBLIC_API_URL ?? 'http://localhost:8080';

type Priority = 'critical' | 'high' | 'medium' | 'low';
type ReviewStatus = 'auto_accepted' | 'needs_review' | 'rejected';

type Detection = {
  id: string;
  label: string;
  normalized_object_type: string;
  confidence: number;
  review_status: ReviewStatus;
  priority: Priority;
  reason: string;
};

type AnalysisResult = {
  analysis_id: string;
  model_mode: string;
  model_id: string;
  kvkk_safe: boolean;
  detections: Detection[];
  created_at: string;
};

type MaintenanceReport = {
  summary: string;
  recommended_action: string;
  risk_level: string;
  kvkk_note: string;
};

type ReviewItem = {
  analysis_id: string;
  detection_id: string;
  label: string;
  normalized_object_type: string;
  confidence: number;
  priority: string;
  reason: string;
};

type SummaryResponse = {
  analysis_id: string;
  report: MaintenanceReport | null;
};

function formatConfidence(confidence: number): string {
  const pct = confidence <= 1 ? confidence * 100 : confidence;
  return `${Math.round(pct)}%`;
}

function getPriorityStyle(priority: string): { chip: object; text: object } {
  const normalized = priority.toLowerCase();
  if (normalized === 'critical' || normalized === 'high') {
    return {
      chip: styles.priorityWarm,
      text: styles.priorityWarmText,
    };
  }
  if (normalized === 'medium') {
    return {
      chip: styles.priorityMedium,
      text: styles.priorityMediumText,
    };
  }
  return {
    chip: styles.priorityLow,
    text: styles.priorityLowText,
  };
}

function formatDate(iso: string): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) {
    return iso;
  }
  return date.toLocaleString('tr-TR');
}

export default function App() {
  const [latestAnalysis, setLatestAnalysis] = useState<AnalysisResult | null>(null);
  const [reviews, setReviews] = useState<ReviewItem[]>([]);
  const [report, setReport] = useState<MaintenanceReport | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    try {
      const [demoRes, reviewsRes, summaryRes] = await Promise.all([
        fetch(`${API_URL}/api/v1/vision/demo-results`),
        fetch(`${API_URL}/api/v1/vision/reviews`),
        fetch(`${API_URL}/api/v1/vision/summary`),
      ]);

      if (!demoRes.ok || !reviewsRes.ok) {
        throw new Error('API request failed');
      }

      const demoResults = (await demoRes.json()) as AnalysisResult[];
      const reviewItems = (await reviewsRes.json()) as ReviewItem[];

      const latest =
        demoResults.length > 0 ? demoResults[demoResults.length - 1] : null;

      setLatestAnalysis(latest);
      setReviews(reviewItems);
      setError(null);

      if (summaryRes.ok) {
        const summaryData = (await summaryRes.json()) as SummaryResponse;
        setReport(summaryData.report);
      } else {
        setReport(null);
      }
    } catch {
      setLatestAnalysis(null);
      setReviews([]);
      setReport(null);
      setError('Backend erişilemiyor. cd backend && go run .');
    }
  }, []);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      await fetchData();
      if (!cancelled) {
        setLoading(false);
      }
    }

    load();
    return () => {
      cancelled = true;
    };
  }, [fetchData]);

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await fetchData();
    setRefreshing(false);
  }, [fetchData]);

  return (
    <View style={styles.container}>
      <ScrollView
        contentContainerStyle={styles.scrollContent}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
        }
      >
        <Text style={styles.eyebrow}>Belediye Saha Görünümü</Text>
        <Text style={styles.title}>CivicLens Saha</Text>
        <Text style={styles.subtitle}>
          Son analiz ve inceleme bekleyen tespitler
        </Text>

        {loading ? (
          <View style={styles.centered}>
            <ActivityIndicator size="large" color="#1565c0" />
            <Text style={styles.loadingText}>Veriler yükleniyor…</Text>
          </View>
        ) : error ? (
          <View style={styles.errorCard}>
            <Text style={styles.errorTitle}>Bağlantı hatası</Text>
            <Text style={styles.errorText}>{error}</Text>
            <Text style={styles.apiUrl}>API: {API_URL}</Text>
          </View>
        ) : (
          <>
            <View style={styles.card}>
              <Text style={styles.cardTitle}>Son Analiz</Text>
              {latestAnalysis ? (
                <>
                  <View style={styles.row}>
                    <Text style={styles.label}>Model modu</Text>
                    <Text style={styles.value}>{latestAnalysis.model_mode}</Text>
                  </View>
                  <View style={styles.row}>
                    <Text style={styles.label}>Model</Text>
                    <Text style={styles.value}>{latestAnalysis.model_id}</Text>
                  </View>
                  <View style={styles.row}>
                    <Text style={styles.label}>KVKK güvenli</Text>
                    <Text
                      style={
                        latestAnalysis.kvkk_safe ? styles.kvkkOk : styles.kvkkWarn
                      }
                    >
                      {latestAnalysis.kvkk_safe ? 'Evet' : 'Hayır'}
                    </Text>
                  </View>
                  <View style={styles.row}>
                    <Text style={styles.label}>Tespit sayısı</Text>
                    <Text style={styles.value}>
                      {latestAnalysis.detections.length}
                    </Text>
                  </View>
                  <Text style={styles.meta}>
                    {formatDate(latestAnalysis.created_at)}
                  </Text>
                </>
              ) : (
                <Text style={styles.emptyText}>Henüz analiz kaydı yok.</Text>
              )}
            </View>

            {report ? (
              <View style={styles.card}>
                <Text style={styles.cardTitle}>Bakım Özeti</Text>
                <Text style={styles.reportSummary}>{report.summary}</Text>
                <Text style={styles.reportAction}>{report.recommended_action}</Text>
                <View style={styles.row}>
                  <Text style={styles.label}>Risk</Text>
                  <Text style={styles.value}>{report.risk_level}</Text>
                </View>
                <Text style={styles.kvkkNote}>{report.kvkk_note}</Text>
              </View>
            ) : null}

            <View style={styles.sectionHeader}>
              <Text style={styles.sectionTitle}>İnceleme Bekleyenler</Text>
              <Text style={styles.sectionCount}>{reviews.length} kayıt</Text>
            </View>

            {reviews.length === 0 ? (
              <View style={styles.card}>
                <Text style={styles.emptyText}>
                  İnceleme bekleyen tespit bulunmuyor.
                </Text>
              </View>
            ) : (
              reviews.map((item) => {
                const priorityStyle = getPriorityStyle(item.priority);
                return (
                  <View
                    key={`${item.analysis_id}-${item.detection_id}`}
                    style={styles.reviewCard}
                  >
                    <View style={styles.reviewHeader}>
                      <Text style={styles.objectType}>
                        {item.normalized_object_type}
                      </Text>
                      <View style={[styles.priorityChip, priorityStyle.chip]}>
                        <Text style={[styles.priorityText, priorityStyle.text]}>
                          {item.priority}
                        </Text>
                      </View>
                    </View>
                    <Text style={styles.reviewLabel}>{item.label}</Text>
                    <Text style={styles.confidence}>
                      Güven: {formatConfidence(item.confidence)}
                    </Text>
                    <Text style={styles.reason}>{item.reason}</Text>
                  </View>
                );
              })
            )}
          </>
        )}
      </ScrollView>
      <StatusBar style="auto" />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
  },
  scrollContent: {
    padding: 20,
    paddingTop: 56,
    paddingBottom: 32,
  },
  eyebrow: {
    fontSize: 12,
    fontWeight: '600',
    textTransform: 'uppercase',
    letterSpacing: 1,
    color: '#666',
    marginBottom: 6,
  },
  title: {
    fontSize: 28,
    fontWeight: '700',
    color: '#111',
    marginBottom: 4,
  },
  subtitle: {
    fontSize: 15,
    lineHeight: 21,
    color: '#555',
    marginBottom: 20,
  },
  centered: {
    alignItems: 'center',
    paddingVertical: 48,
    gap: 12,
  },
  loadingText: {
    fontSize: 14,
    color: '#666',
  },
  errorCard: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 20,
    borderWidth: 1,
    borderColor: '#f5c2c7',
  },
  errorTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#b00020',
    marginBottom: 8,
  },
  errorText: {
    fontSize: 15,
    lineHeight: 22,
    color: '#b00020',
    marginBottom: 12,
  },
  apiUrl: {
    fontSize: 12,
    color: '#888',
  },
  card: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 16,
    borderWidth: 1,
    borderColor: '#e5e5e5',
    marginBottom: 16,
  },
  cardTitle: {
    fontSize: 17,
    fontWeight: '700',
    color: '#111',
    marginBottom: 12,
  },
  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  label: {
    fontSize: 14,
    color: '#666',
  },
  value: {
    fontSize: 14,
    fontWeight: '600',
    color: '#222',
    flexShrink: 1,
    textAlign: 'right',
    marginLeft: 12,
  },
  kvkkOk: {
    fontSize: 14,
    fontWeight: '700',
    color: '#0a7a2f',
  },
  kvkkWarn: {
    fontSize: 14,
    fontWeight: '700',
    color: '#b00020',
  },
  meta: {
    fontSize: 12,
    color: '#888',
    marginTop: 4,
  },
  emptyText: {
    fontSize: 14,
    color: '#666',
    lineHeight: 20,
  },
  reportSummary: {
    fontSize: 14,
    lineHeight: 20,
    color: '#333',
    marginBottom: 8,
  },
  reportAction: {
    fontSize: 14,
    lineHeight: 20,
    color: '#1565c0',
    marginBottom: 8,
  },
  kvkkNote: {
    fontSize: 12,
    lineHeight: 18,
    color: '#888',
    marginTop: 4,
  },
  sectionHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'baseline',
    marginBottom: 10,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: '700',
    color: '#111',
  },
  sectionCount: {
    fontSize: 13,
    color: '#666',
  },
  reviewCard: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 14,
    borderWidth: 1,
    borderColor: '#e5e5e5',
    marginBottom: 10,
  },
  reviewHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 6,
  },
  objectType: {
    fontSize: 15,
    fontWeight: '700',
    color: '#222',
    flex: 1,
    marginRight: 8,
  },
  priorityChip: {
    borderRadius: 999,
    paddingHorizontal: 10,
    paddingVertical: 4,
  },
  priorityText: {
    fontSize: 11,
    fontWeight: '700',
    textTransform: 'uppercase',
  },
  priorityWarm: {
    backgroundColor: '#fde8e8',
  },
  priorityWarmText: {
    color: '#c62828',
  },
  priorityMedium: {
    backgroundColor: '#e3f2fd',
  },
  priorityMediumText: {
    color: '#1565c0',
  },
  priorityLow: {
    backgroundColor: '#eeeeee',
  },
  priorityLowText: {
    color: '#616161',
  },
  reviewLabel: {
    fontSize: 13,
    color: '#555',
    marginBottom: 4,
  },
  confidence: {
    fontSize: 13,
    fontWeight: '600',
    color: '#333',
    marginBottom: 6,
  },
  reason: {
    fontSize: 13,
    lineHeight: 19,
    color: '#444',
  },
});
