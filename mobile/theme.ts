import { StyleSheet } from 'react-native';

export function getPriorityStyle(priority: string): {
  chip: object;
  text: object;
} {
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

export function getStatusStyle(status: string): {
  chip: object;
  text: object;
} {
  const normalized = status.toLowerCase();
  if (
    normalized === 'completed' ||
    normalized === 'ai_verified' ||
    normalized === 'likely_resolved'
  ) {
    return { chip: styles.statusOk, text: styles.statusOkText };
  }
  if (
    normalized === 'assigned' ||
    normalized === 'started' ||
    normalized === 'waiting_for_review'
  ) {
    return { chip: styles.statusActive, text: styles.statusActiveText };
  }
  if (normalized === 'reopened' || normalized === 'rejected') {
    return { chip: styles.statusWarn, text: styles.statusWarnText };
  }
  return { chip: styles.statusNeutral, text: styles.statusNeutralText };
}

export const sharedStyles = StyleSheet.create({
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
    marginBottom: 16,
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
  emptyText: {
    fontSize: 14,
    color: '#666',
    lineHeight: 20,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: '700',
    color: '#111',
    marginBottom: 10,
  },
  chip: {
    borderRadius: 999,
    paddingHorizontal: 10,
    paddingVertical: 4,
  },
  chipText: {
    fontSize: 11,
    fontWeight: '700',
    textTransform: 'uppercase',
  },
  input: {
    backgroundColor: '#fafafa',
    borderWidth: 1,
    borderColor: '#ddd',
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 10,
    fontSize: 15,
    color: '#222',
    marginBottom: 12,
  },
  inputMultiline: {
    minHeight: 88,
    textAlignVertical: 'top',
  },
  button: {
    backgroundColor: '#1565c0',
    borderRadius: 10,
    paddingVertical: 14,
    alignItems: 'center',
  },
  buttonDisabled: {
    opacity: 0.6,
  },
  buttonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '700',
  },
  buttonSecondary: {
    backgroundColor: '#e8f0fe',
    borderRadius: 10,
    paddingVertical: 10,
    paddingHorizontal: 14,
    alignItems: 'center',
    marginTop: 8,
  },
  buttonSecondaryText: {
    color: '#1565c0',
    fontSize: 14,
    fontWeight: '700',
  },
  successText: {
    fontSize: 14,
    color: '#0a7a2f',
    lineHeight: 20,
    marginTop: 8,
  },
  hintText: {
    fontSize: 13,
    color: '#888',
    lineHeight: 18,
    marginBottom: 12,
  },
});

const styles = StyleSheet.create({
  priorityWarm: { backgroundColor: '#fde8e8' },
  priorityWarmText: { color: '#c62828' },
  priorityMedium: { backgroundColor: '#e3f2fd' },
  priorityMediumText: { color: '#1565c0' },
  priorityLow: { backgroundColor: '#eeeeee' },
  priorityLowText: { color: '#616161' },
  statusOk: { backgroundColor: '#e8f5e9' },
  statusOkText: { color: '#2e7d32' },
  statusActive: { backgroundColor: '#e3f2fd' },
  statusActiveText: { color: '#1565c0' },
  statusWarn: { backgroundColor: '#fff3e0' },
  statusWarnText: { color: '#e65100' },
  statusNeutral: { backgroundColor: '#eeeeee' },
  statusNeutralText: { color: '#616161' },
});
