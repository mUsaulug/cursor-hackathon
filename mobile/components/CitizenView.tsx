import { useCallback, useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from 'react-native';

import {
  API_URL,
  BACKEND_UNREACHABLE,
  apiFetch,
  appendFile,
  getPlaceholderImage,
} from '../api';
import { getPriorityStyle, sharedStyles } from '../theme';
import type { AnalysisResult, Report, Role } from '../types';

type Props = {
  role: Role;
};

function formatConfidence(confidence: number): string {
  const pct = confidence <= 1 ? confidence * 100 : confidence;
  return `${Math.round(pct)}%`;
}

function formatDate(iso: string): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) {
    return iso;
  }
  return date.toLocaleString('tr-TR');
}

export function CitizenView({ role }: Props) {
  const [description, setDescription] = useState('');
  const [lat, setLat] = useState('41.0082');
  const [lng, setLng] = useState('28.9784');
  const [latestAnalysis, setLatestAnalysis] = useState<AnalysisResult | null>(null);
  const [submittedReport, setSubmittedReport] = useState<Report | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [photoFallback, setPhotoFallback] = useState(false);

  const fetchDemo = useCallback(async () => {
    try {
      const res = await apiFetch('/api/v1/vision/demo-results', role);
      if (!res.ok) {
        throw new Error('demo failed');
      }
      const demoResults = (await res.json()) as AnalysisResult[];
      const latest =
        demoResults.length > 0 ? demoResults[demoResults.length - 1] : null;
      setLatestAnalysis(latest);
      setError(null);
    } catch {
      setLatestAnalysis(null);
      setError(BACKEND_UNREACHABLE);
    }
  }, [role]);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      await fetchDemo();
      if (!cancelled) {
        setLoading(false);
      }
    }

    setLoading(true);
    load();
    return () => {
      cancelled = true;
    };
  }, [fetchDemo]);

  const handleSubmit = useCallback(async () => {
    if (!description.trim()) {
      Alert.alert('Eksik bilgi', 'Lütfen bir açıklama girin.');
      return;
    }

    const imageFile = getPlaceholderImage();
    if (!imageFile) {
      setPhotoFallback(true);
      Alert.alert(
        'Foto yükleme',
        'Foto yükleme cihazda kamerayla yapılır. Yerel görsel bulunamadı.',
      );
      return;
    }

    setSubmitting(true);
    setSubmittedReport(null);
    setPhotoFallback(false);

    try {
      const formData = new FormData();
      formData.append('description', description.trim());
      formData.append('lat', lat);
      formData.append('lng', lng);
      appendFile(formData, 'image', imageFile);

      const res = await apiFetch('/api/v1/reports', role, {
        method: 'POST',
        body: formData,
      });

      if (!res.ok) {
        const body = (await res.json().catch(() => null)) as { error?: string } | null;
        throw new Error(body?.error ?? 'Bildirim gönderilemedi');
      }

      const report = (await res.json()) as Report;
      setSubmittedReport(report);
      setDescription('');
      setError(null);
      Alert.alert('Başarılı', `Bildirim oluşturuldu: ${report.report_id}`);
    } catch (err) {
      const message =
        err instanceof Error && err.message !== 'Failed to fetch'
          ? err.message
          : BACKEND_UNREACHABLE;
      setError(message);
    } finally {
      setSubmitting(false);
    }
  }, [description, lat, lng, role]);

  if (loading) {
    return (
      <View style={sharedStyles.centered}>
        <ActivityIndicator size="large" color="#1565c0" />
        <Text style={sharedStyles.loadingText}>Veriler yükleniyor…</Text>
      </View>
    );
  }

  return (
    <ScrollView
      style={styles.scroll}
      contentContainerStyle={styles.scrollContent}
    >
      {error ? (
        <View style={sharedStyles.errorCard}>
          <Text style={sharedStyles.errorTitle}>Bağlantı hatası</Text>
          <Text style={sharedStyles.errorText}>{error}</Text>
          <Text style={sharedStyles.apiUrl}>API: {API_URL}</Text>
        </View>
      ) : null}

      <View style={sharedStyles.card}>
        <Text style={sharedStyles.cardTitle}>AI Analiz Demosu</Text>
        {latestAnalysis ? (
          <>
            <View style={sharedStyles.row}>
              <Text style={sharedStyles.label}>Model</Text>
              <Text style={sharedStyles.value}>{latestAnalysis.model_id}</Text>
            </View>
            <View style={sharedStyles.row}>
              <Text style={sharedStyles.label}>KVKK güvenli</Text>
              <Text
                style={
                  latestAnalysis.kvkk_safe ? styles.kvkkOk : styles.kvkkWarn
                }
              >
                {latestAnalysis.kvkk_safe ? 'Evet' : 'Hayır'}
              </Text>
            </View>
            <View style={sharedStyles.row}>
              <Text style={sharedStyles.label}>Tespit sayısı</Text>
              <Text style={sharedStyles.value}>
                {latestAnalysis.detections.length}
              </Text>
            </View>
            {latestAnalysis.detections.slice(0, 3).map((det) => {
              const priorityStyle = getPriorityStyle(det.priority);
              return (
                <View key={det.id} style={styles.detectionRow}>
                  <Text style={styles.detectionLabel}>
                    {det.normalized_object_type}
                  </Text>
                  <View style={[sharedStyles.chip, priorityStyle.chip]}>
                    <Text style={[sharedStyles.chipText, priorityStyle.text]}>
                      {det.priority}
                    </Text>
                  </View>
                  <Text style={styles.detectionMeta}>
                    Güven: {formatConfidence(det.confidence)}
                  </Text>
                </View>
              );
            })}
            <Text style={styles.meta}>
              {formatDate(latestAnalysis.created_at)}
            </Text>
          </>
        ) : (
          <Text style={sharedStyles.emptyText}>Henüz analiz kaydı yok.</Text>
        )}
      </View>

      <View style={sharedStyles.card}>
        <Text style={sharedStyles.cardTitle}>Sorun Bildir</Text>
        <Text style={sharedStyles.hintText}>
          Fotoğraf için uygulama simgesi kullanılır (demo). Gerçek cihazda kamera
          ile çekim yapılır.
        </Text>

        {photoFallback ? (
          <Text style={styles.fallbackText}>
            Foto yükleme cihazda kamerayla yapılır
          </Text>
        ) : null}

        <Text style={styles.fieldLabel}>Açıklama</Text>
        <TextInput
          style={[sharedStyles.input, sharedStyles.inputMultiline]}
          value={description}
          onChangeText={setDescription}
          placeholder="Örn: Kaldırımda çukur var"
          placeholderTextColor="#999"
          multiline
        />

        <Text style={styles.fieldLabel}>Enlem (lat)</Text>
        <TextInput
          style={sharedStyles.input}
          value={lat}
          onChangeText={setLat}
          keyboardType="decimal-pad"
          placeholder="41.0082"
          placeholderTextColor="#999"
        />

        <Text style={styles.fieldLabel}>Boylam (lng)</Text>
        <TextInput
          style={sharedStyles.input}
          value={lng}
          onChangeText={setLng}
          keyboardType="decimal-pad"
          placeholder="28.9784"
          placeholderTextColor="#999"
        />

        <Pressable
          style={[sharedStyles.button, submitting && sharedStyles.buttonDisabled]}
          onPress={handleSubmit}
          disabled={submitting}
        >
          {submitting ? (
            <ActivityIndicator color="#fff" />
          ) : (
            <Text style={sharedStyles.buttonText}>Bildir</Text>
          )}
        </Pressable>

        {submittedReport ? (
          <View style={styles.resultBox}>
            <Text style={sharedStyles.successText}>
              Bildirim #{submittedReport.report_id} alındı
            </Text>
            <View style={sharedStyles.row}>
              <Text style={sharedStyles.label}>Durum</Text>
              <Text style={sharedStyles.value}>{submittedReport.status}</Text>
            </View>
            {submittedReport.problem_type ? (
              <View style={sharedStyles.row}>
                <Text style={sharedStyles.label}>Problem tipi</Text>
                <Text style={sharedStyles.value}>
                  {submittedReport.problem_type}
                </Text>
              </View>
            ) : null}
            {submittedReport.priority ? (
              <View style={sharedStyles.row}>
                <Text style={sharedStyles.label}>Öncelik</Text>
                <Text style={sharedStyles.value}>{submittedReport.priority}</Text>
              </View>
            ) : null}
          </View>
        ) : null}
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  scroll: {
    flex: 1,
  },
  scrollContent: {
    paddingBottom: 32,
  },
  fieldLabel: {
    fontSize: 13,
    fontWeight: '600',
    color: '#444',
    marginBottom: 6,
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
  detectionRow: {
    marginBottom: 10,
    paddingBottom: 10,
    borderBottomWidth: 1,
    borderBottomColor: '#f0f0f0',
  },
  detectionLabel: {
    fontSize: 14,
    fontWeight: '600',
    color: '#333',
    marginBottom: 4,
  },
  detectionMeta: {
    fontSize: 12,
    color: '#666',
    marginTop: 4,
  },
  meta: {
    fontSize: 12,
    color: '#888',
    marginTop: 4,
  },
  fallbackText: {
    fontSize: 14,
    color: '#e65100',
    marginBottom: 12,
    lineHeight: 20,
  },
  resultBox: {
    marginTop: 16,
    paddingTop: 12,
    borderTopWidth: 1,
    borderTopColor: '#e5e5e5',
  },
});
