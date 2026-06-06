import { useCallback, useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  FlatList,
  Pressable,
  RefreshControl,
  StyleSheet,
  Text,
  View,
} from 'react-native';

import {
  API_URL,
  BACKEND_UNREACHABLE,
  apiFetch,
  appendFile,
  getPlaceholderImage,
} from '../api';
import { getPriorityStyle, getStatusStyle, sharedStyles } from '../theme';
import type { Evidence, Role, Task } from '../types';

const ASSIGNED_TO = 'saha_ekip_1';

type Props = {
  role: Role;
};

type ListItem =
  | { type: 'header'; key: string; title: string }
  | { type: 'task'; key: string; task: Task };

function buildListItems(assigned: Task[], all: Task[]): ListItem[] {
  const items: ListItem[] = [
    { type: 'header', key: 'hdr-assigned', title: 'Bana Atanan Görevler' },
    ...assigned.map((task) => ({
      type: 'task' as const,
      key: `assigned-${task.task_id}`,
      task,
    })),
    { type: 'header', key: 'hdr-all', title: 'Tüm Görevler' },
    ...all.map((task) => ({
      type: 'task' as const,
      key: `all-${task.task_id}`,
      task,
    })),
  ];
  return items;
}

function TaskCard({
  task,
  onAction,
  busyTaskId,
  lastEvidence,
}: {
  task: Task;
  onAction: (task: Task, action: 'start' | 'evidence') => void;
  busyTaskId: string | null;
  lastEvidence: Evidence | null;
}) {
  const priorityStyle = getPriorityStyle(task.priority);
  const statusStyle = getStatusStyle(task.status);
  const busy = busyTaskId === task.task_id;
  const showEvidence =
    lastEvidence !== null && lastEvidence.task_id === task.task_id;

  return (
    <View style={styles.taskCard}>
      <View style={styles.taskHeader}>
        <Text style={styles.taskId}>{task.task_id}</Text>
        <View style={[sharedStyles.chip, priorityStyle.chip]}>
          <Text style={[sharedStyles.chipText, priorityStyle.text]}>
            {task.priority}
          </Text>
        </View>
      </View>

      <View style={sharedStyles.row}>
        <Text style={sharedStyles.label}>Müdürlük</Text>
        <Text style={sharedStyles.value}>{task.assigned_department}</Text>
      </View>
      <View style={sharedStyles.row}>
        <Text style={sharedStyles.label}>Atanan</Text>
        <Text style={sharedStyles.value}>{task.assigned_to || '—'}</Text>
      </View>
      <View style={sharedStyles.row}>
        <Text style={sharedStyles.label}>SLA</Text>
        <Text style={sharedStyles.value}>{task.sla || '—'}</Text>
      </View>

      <View style={styles.statusRow}>
        <Text style={sharedStyles.label}>Durum</Text>
        <View style={[sharedStyles.chip, statusStyle.chip]}>
          <Text style={[sharedStyles.chipText, statusStyle.text]}>
            {task.status}
          </Text>
        </View>
      </View>

      {task.status === 'assigned' ? (
        <Pressable
          style={[sharedStyles.buttonSecondary, busy && sharedStyles.buttonDisabled]}
          onPress={() => onAction(task, 'start')}
          disabled={busy}
        >
          {busy ? (
            <ActivityIndicator color="#1565c0" />
          ) : (
            <Text style={sharedStyles.buttonSecondaryText}>Başlat</Text>
          )}
        </Pressable>
      ) : null}

      {task.status === 'started' ? (
        <Pressable
          style={[sharedStyles.buttonSecondary, busy && sharedStyles.buttonDisabled]}
          onPress={() => onAction(task, 'evidence')}
          disabled={busy}
        >
          {busy ? (
            <ActivityIndicator color="#1565c0" />
          ) : (
            <Text style={sharedStyles.buttonSecondaryText}>Kanıt Yükle</Text>
          )}
        </Pressable>
      ) : null}

      {showEvidence && lastEvidence ? (
        <View style={styles.evidenceBox}>
          <Text style={styles.evidenceTitle}>AI Doğrulama</Text>
          <Text style={styles.evidenceValue}>{lastEvidence.ai_verification}</Text>
          <Text style={styles.evidenceMeta}>
            Yönetici onayı: {lastEvidence.manager_approval}
          </Text>
        </View>
      ) : null}
    </View>
  );
}

export function FieldStaffView({ role }: Props) {
  const [assignedTasks, setAssignedTasks] = useState<Task[]>([]);
  const [allTasks, setAllTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [busyTaskId, setBusyTaskId] = useState<string | null>(null);
  const [lastEvidence, setLastEvidence] = useState<Evidence | null>(null);

  const fetchTasks = useCallback(async () => {
    try {
      const [assignedRes, allRes] = await Promise.all([
        apiFetch(`/api/v1/tasks?assigned_to=${ASSIGNED_TO}`, role),
        apiFetch('/api/v1/tasks', role),
      ]);

      if (!assignedRes.ok || !allRes.ok) {
        throw new Error('tasks failed');
      }

      const assigned = (await assignedRes.json()) as Task[];
      const all = (await allRes.json()) as Task[];
      setAssignedTasks(assigned);
      setAllTasks(all);
      setError(null);
    } catch {
      setAssignedTasks([]);
      setAllTasks([]);
      setError(BACKEND_UNREACHABLE);
    }
  }, [role]);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      await fetchTasks();
      if (!cancelled) {
        setLoading(false);
      }
    }

    setLoading(true);
    load();
    return () => {
      cancelled = true;
    };
  }, [fetchTasks]);

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await fetchTasks();
    setRefreshing(false);
  }, [fetchTasks]);

  const handleAction = useCallback(
    async (task: Task, action: 'start' | 'evidence') => {
      setBusyTaskId(task.task_id);

      try {
        if (action === 'start') {
          const res = await apiFetch(`/api/v1/tasks/${task.task_id}/start`, role, {
            method: 'POST',
          });
          if (!res.ok) {
            const body = (await res.json().catch(() => null)) as {
              error?: string;
            } | null;
            throw new Error(body?.error ?? 'Görev başlatılamadı');
          }
          Alert.alert('Başarılı', 'Görev başlatıldı.');
          await fetchTasks();
          return;
        }

        const imageFile = getPlaceholderImage();
        if (!imageFile) {
          Alert.alert(
            'Foto yükleme',
            'Foto yükleme cihazda kamerayla yapılır. Yerel görsel bulunamadı.',
          );
          return;
        }

        const formData = new FormData();
        appendFile(formData, 'image', imageFile);
        formData.append('uploaded_by', ASSIGNED_TO);

        const res = await apiFetch(
          `/api/v1/tasks/${task.task_id}/evidence`,
          role,
          {
            method: 'POST',
            body: formData,
          },
        );

        if (!res.ok) {
          const body = (await res.json().catch(() => null)) as {
            error?: string;
          } | null;
          throw new Error(body?.error ?? 'Kanıt yüklenemedi');
        }

        const evidence = (await res.json()) as Evidence;
        setLastEvidence(evidence);
        Alert.alert(
          'Kanıt yüklendi',
          `AI doğrulama: ${evidence.ai_verification}`,
        );
        await fetchTasks();
      } catch (err) {
        const message =
          err instanceof Error && err.message !== 'Failed to fetch'
            ? err.message
            : BACKEND_UNREACHABLE;
        Alert.alert('Hata', message);
        if (message === BACKEND_UNREACHABLE) {
          setError(message);
        }
      } finally {
        setBusyTaskId(null);
      }
    },
    [fetchTasks, role],
  );

  const listItems = buildListItems(assignedTasks, allTasks);

  if (loading) {
    return (
      <View style={sharedStyles.centered}>
        <ActivityIndicator size="large" color="#1565c0" />
        <Text style={sharedStyles.loadingText}>Görevler yükleniyor…</Text>
      </View>
    );
  }

  return (
    <FlatList
      style={styles.list}
      data={listItems}
      keyExtractor={(item) => item.key}
      contentContainerStyle={styles.listContent}
      refreshControl={
        <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
      }
      ListHeaderComponent={
        error ? (
          <View style={sharedStyles.errorCard}>
            <Text style={sharedStyles.errorTitle}>Bağlantı hatası</Text>
            <Text style={sharedStyles.errorText}>{error}</Text>
            <Text style={sharedStyles.apiUrl}>API: {API_URL}</Text>
          </View>
        ) : null
      }
      ListEmptyComponent={
        !error ? (
          <View style={sharedStyles.card}>
            <Text style={sharedStyles.emptyText}>Görev bulunmuyor.</Text>
          </View>
        ) : null
      }
      renderItem={({ item }) => {
        if (item.type === 'header') {
          return (
            <Text style={[sharedStyles.sectionTitle, styles.sectionHeader]}>
              {item.title}
            </Text>
          );
        }

        return (
          <TaskCard
            task={item.task}
            onAction={handleAction}
            busyTaskId={busyTaskId}
            lastEvidence={lastEvidence}
          />
        );
      }}
    />
  );
}

const styles = StyleSheet.create({
  list: {
    flex: 1,
  },
  listContent: {
    paddingBottom: 32,
  },
  sectionHeader: {
    marginTop: 8,
  },
  taskCard: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 14,
    borderWidth: 1,
    borderColor: '#e5e5e5',
    marginBottom: 10,
  },
  taskHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 10,
  },
  taskId: {
    fontSize: 15,
    fontWeight: '700',
    color: '#222',
    flex: 1,
    marginRight: 8,
  },
  statusRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 4,
    marginTop: 4,
  },
  evidenceBox: {
    marginTop: 12,
    padding: 12,
    backgroundColor: '#f0f7ff',
    borderRadius: 8,
  },
  evidenceTitle: {
    fontSize: 13,
    fontWeight: '700',
    color: '#1565c0',
    marginBottom: 4,
  },
  evidenceValue: {
    fontSize: 15,
    fontWeight: '600',
    color: '#222',
    marginBottom: 4,
  },
  evidenceMeta: {
    fontSize: 12,
    color: '#666',
  },
});
