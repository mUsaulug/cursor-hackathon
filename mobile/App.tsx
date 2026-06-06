import { useState } from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';

import { CitizenView } from './components/CitizenView';
import { FieldStaffView } from './components/FieldStaffView';
import { API_URL } from './api';
import type { Role } from './types';

type RoleOption = {
  role: Role;
  label: string;
};

const ROLE_OPTIONS: RoleOption[] = [
  { role: 'citizen', label: 'Vatandaş' },
  { role: 'field_staff', label: 'Saha' },
];

export default function App() {
  const [role, setRole] = useState<Role>('citizen');

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
        {role === 'citizen' ? (
          <CitizenView role={role} />
        ) : (
          <FieldStaffView role={role} />
        )}
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
    paddingBottom: 24,
  },
});
