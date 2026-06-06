import { useCallback, useState } from 'react';
import {
  ActivityIndicator,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  View,
} from 'react-native';

import { CitizenView } from './components/CitizenView';
import { FieldStaffView } from './components/FieldStaffView';
import { API_URL } from './api';
import { sharedStyles } from './theme';
import type { Report, Role } from './types';

type RoleOption = {
  role: Role;
  label: string;
};

const ROLE_OPTIONS: RoleOption[] = [
  { role: 'citizen', label: 'Vatandaş' },
  { role: 'field_staff', label: 'Saha' },
];

type ActiveTab = 'analysis' | 'report';

type SubmitState =
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'success'; reportId: string }
  | { status: 'error'; message: string };

function ReportTab() {
  const [description, setDescription] = useState('');
  const [lat, setLat] = useState('');
  const [lng, setLng] = useState('');
  const [submitState, setSubmitState] = useState<SubmitState>({ status: 'idle' });

  const handleSubmit = useCallback(async () => {
    if (!description.trim()) {
      setSubmitState({ status: 'error', message: 'Açıklama alanı zorunludur.' });
      return;
    }

    setSubmitState({ status: 'loading' });

    try {
      const body = new FormData();
      body.append('description', description.trim());
      body.append('source_type', 'citizen_mobile');
      if (lat.trim()) {
        body.append('lat', lat.trim());
      }
      if (lng.trim()) {
        body.append('lng', lng.trim());
      }

      const res = await fetch(`${API_URL}/api/v1/reports`, {
        method: 'POST',
        headers: { 'X-Role': 'citizen' },
        body,
      });

      if (!res.ok) {
        const payload = (await res.json().catch(() => null)) as { error?: string } | null;
        throw new Error(payload?.error ?? `Sunucu hatası: ${res.status}`);
      }

      const report = (await res.json()) as Report;
      setSubmitState({ status: 'success', reportId: report.report_id });
      setDescription('');
      setLat('');
      setLng('');
    } catch (err: unknown) {
      const message =
        err instanceof Error ? err.message : 'Bilinmeyen bir hata oluştu.';
      setSubmitState({ status: 'error', message });
    }
  }, [description, lat, lng]);

  const isLoading = submitState.status === 'loading';

  return (
    <ScrollView
      style={tabStyles.scroll}
      contentContainerStyle={tabStyles.scrollContent}
      keyboardShouldPersistTaps="handled"
    >
      {submitState.status === 'success' && (
        <View style={tabStyles.successCard}>
          <Text style={tabStyles.successText}>
            Bildiriminiz gönderildi!{'\n'}
            <Text style={tabStyles.successId}>#{submitState.reportId}</Text>
          </Text>
        </View>
      )}

      {submitState.status === 'error' && (
        <View style={tabStyles.errorCard}>
          <Text style={tabStyles.errorText}>{submitState.message}</Text>
        </View>
      )}

      <View style={sharedStyles.card}>
        <Text style={sharedStyles.cardTitle}>Sorun Bildir</Text>

        <Text style={tabStyles.fieldLabel}>Açıklama *</Text>
        <TextInput
          style={[sharedStyles.input, sharedStyles.inputMultiline]}
          value={description}
          onChangeText={(text) => setDescription(text)}
          placeholder="Örn: Kaldırımda çukur var"
          placeholderTextColor="#999"
          multiline
          editable={!isLoading}
        />

        <Text style={tabStyles.fieldLabel}>Konum (opsiyonel)</Text>
        <View style={tabStyles.coordRow}>
          <TextInput
            style={[sharedStyles.input, tabStyles.coordInput]}
            value={lat}
            onChangeText={(text) => setLat(text)}
            keyboardType="decimal-pad"
            placeholder="Enlem"
            placeholderTextColor="#999"
            editable={!isLoading}
          />
          <TextInput
            style={[sharedStyles.input, tabStyles.coordInput]}
            value={lng}
            onChangeText={(text) => setLng(text)}
            keyboardType="decimal-pad"
            placeholder="Boylam"
            placeholderTextColor="#999"
            editable={!isLoading}
          />
        </View>

        <Pressable
          style={[sharedStyles.button, isLoading && sharedStyles.buttonDisabled]}
          onPress={handleSubmit}
          disabled={isLoading}
        >
          {isLoading ? (
            <ActivityIndicator color="#fff" />
          ) : (
            <Text style={sharedStyles.buttonText}>Bildirimi Gönder</Text>
          )}
        </Pressable>
      </View>
    </ScrollView>
  );
}

export default function App() {
  const [role, setRole] = useState<Role>('citizen');
  const [activeTab, setActiveTab] = useState<ActiveTab>('analysis');

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.eyebrow}>CivicLens Wave 2</Text>
        <Text style={styles.title}>Saha Uygulaması</Text>
        <Text style={styles.subtitle}>
          Rol seçin; tüm istekler X-Role başlığı ile gönderilir.
        </Text>
        <Text style={styles.apiHint}>API: {API_URL}</Text>

        <View style={styles.roleRow}>
          {ROLE_OPTIONS.map((option) => {
            const selected = role === option.role;
            return (
              <Pressable
                key={option.role}
                style={[styles.roleChip, selected && styles.roleChipSelected]}
                onPress={() => setRole(option.role)}
              >
                <Text
                  style={[
                    styles.roleChipText,
                    selected && styles.roleChipTextSelected,
                  ]}
                >
                  {option.label}
                </Text>
              </Pressable>
            );
          })}
        </View>
      </View>

      <View style={styles.content}>
        {activeTab === 'analysis' ? (
          role === 'citizen' ? (
            <CitizenView role={role} />
          ) : (
            <FieldStaffView role={role} />
          )
        ) : (
          <ReportTab />
        )}
      </View>

      {/* Bottom Tab Bar */}
      <View style={styles.tabBar}>
        <TouchableOpacity
          style={styles.tabItem}
          onPress={() => setActiveTab('analysis')}
          activeOpacity={0.7}
        >
          <View
            style={[
              styles.tabIndicator,
              activeTab === 'analysis' && styles.tabIndicatorActive,
            ]}
          />
          <Text
            style={[
              styles.tabLabel,
              activeTab === 'analysis' && styles.tabLabelActive,
            ]}
          >
            Analiz
          </Text>
        </TouchableOpacity>

        <TouchableOpacity
          style={styles.tabItem}
          onPress={() => setActiveTab('report')}
          activeOpacity={0.7}
        >
          <View
            style={[
              styles.tabIndicator,
              activeTab === 'report' && styles.tabIndicatorActive,
            ]}
          />
          <Text
            style={[
              styles.tabLabel,
              activeTab === 'report' && styles.tabLabelActive,
            ]}
          >
            Bildirim
          </Text>
        </TouchableOpacity>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f8fafc',
  },
  header: {
    paddingHorizontal: 20,
    paddingTop: 56,
    paddingBottom: 12,
    backgroundColor: '#f8fafc',
    borderBottomWidth: 1,
    borderBottomColor: '#e5e5e5',
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
    marginBottom: 8,
  },
  apiHint: {
    fontSize: 12,
    color: '#888',
    marginBottom: 14,
  },
  roleRow: {
    flexDirection: 'row',
    gap: 10,
  },
  roleChip: {
    paddingHorizontal: 18,
    paddingVertical: 10,
    borderRadius: 999,
    backgroundColor: '#fff',
    borderWidth: 1,
    borderColor: '#ccc',
  },
  roleChipSelected: {
    backgroundColor: '#0f172a',
    borderColor: '#0f172a',
  },
  roleChipText: {
    fontSize: 15,
    fontWeight: '600',
    color: '#333',
  },
  roleChipTextSelected: {
    color: '#fff',
  },
  content: {
    flex: 1,
    paddingHorizontal: 20,
    paddingTop: 12,
  },
  // ── Bottom Tab Bar ──
  tabBar: {
    flexDirection: 'row',
    backgroundColor: '#fff',
    borderTopWidth: 1,
    borderTopColor: '#e2e8f0',
    paddingBottom: 20,
  },
  tabItem: {
    flex: 1,
    alignItems: 'center',
    paddingTop: 8,
    paddingBottom: 4,
  },
  tabIndicator: {
    height: 3,
    width: 32,
    borderRadius: 2,
    backgroundColor: 'transparent',
    marginBottom: 4,
  },
  tabIndicatorActive: {
    backgroundColor: '#0f172a',
  },
  tabLabel: {
    fontSize: 13,
    fontWeight: '500',
    color: '#94a3b8',
  },
  tabLabelActive: {
    color: '#0f172a',
    fontWeight: '700',
  },
});

// ── ReportTab styles ──
const tabStyles = StyleSheet.create({
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
  coordRow: {
    flexDirection: 'row',
    gap: 10,
  },
  coordInput: {
    flex: 1,
  },
  successCard: {
    backgroundColor: '#dcfce7',
    borderRadius: 10,
    padding: 14,
    marginBottom: 16,
    borderWidth: 1,
    borderColor: '#86efac',
  },
  successText: {
    fontSize: 15,
    fontWeight: '600',
    color: '#16a34a',
    textAlign: 'center',
    lineHeight: 22,
  },
  successId: {
    fontSize: 13,
    fontWeight: '400',
    color: '#166534',
  },
  errorCard: {
    backgroundColor: '#fee2e2',
    borderRadius: 10,
    padding: 14,
    marginBottom: 16,
    borderWidth: 1,
    borderColor: '#fca5a5',
  },
  errorText: {
    fontSize: 14,
    color: '#dc2626',
    lineHeight: 20,
  },
});
