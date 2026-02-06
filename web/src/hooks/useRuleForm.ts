import { useState, useEffect, useCallback } from 'react';
import { Rule, TargetLayer, Trigger, EnforcementMode, Category } from '@/domain/rule';
import { createRule, updateRule, CreateRuleRequest, fetchCategories } from '@/lib/api';
import { DEFAULT_TIMEOUT_HOURS } from '@/components/RuleEditor/constants';

export interface RuleFormData {
  name: string;
  description: string;
  content: string;
  targetLayer: TargetLayer;
  categoryId: string;
  priorityWeight: number;
  overridable: boolean;
  effectiveStart: string;
  effectiveEnd: string;
  tags: string;
  triggers: Trigger[];
  enforcementMode: EnforcementMode;
  temporaryTimeoutHours: number;
}

export interface RuleFormErrors {
  name?: string;
  content?: string;
  general?: string;
}

interface UseRuleFormOptions {
  teamId: string;
  rule?: Rule;
  onSuccess: () => void;
}

interface UseRuleFormReturn {
  formData: RuleFormData;
  errors: RuleFormErrors;
  categories: Category[];
  saving: boolean;
  setField: <K extends keyof RuleFormData>(field: K, value: RuleFormData[K]) => void;
  setTriggers: (triggers: Trigger[]) => void;
  validate: () => boolean;
  handleSubmit: (e: React.FormEvent) => Promise<void>;
  resetErrors: () => void;
}

function createInitialFormData(rule?: Rule): RuleFormData {
  return {
    name: rule?.name || '',
    description: rule?.description || '',
    content: rule?.content || '',
    targetLayer: rule?.targetLayer || 'project',
    categoryId: rule?.categoryId || '',
    priorityWeight: rule?.priorityWeight || 0,
    overridable: rule?.overridable ?? true,
    effectiveStart: rule?.effectiveStart || '',
    effectiveEnd: rule?.effectiveEnd || '',
    tags: rule?.tags?.join(', ') || '',
    triggers: rule?.triggers || [],
    enforcementMode: rule?.enforcementMode || 'block',
    temporaryTimeoutHours: rule?.temporaryTimeoutHours || DEFAULT_TIMEOUT_HOURS,
  };
}

/**
 * Custom hook for managing RuleEditor form state and validation.
 * Encapsulates form logic following the Controlled Form with Validation pattern.
 */
export function useRuleForm({ teamId, rule, onSuccess }: UseRuleFormOptions): UseRuleFormReturn {
  const [formData, setFormData] = useState<RuleFormData>(() => createInitialFormData(rule));
  const [errors, setErrors] = useState<RuleFormErrors>({});
  const [categories, setCategories] = useState<Category[]>([]);
  const [saving, setSaving] = useState(false);

  // Load categories on mount
  useEffect(() => {
    fetchCategories()
      .then(setCategories)
      .catch(() => setErrors(prev => ({ ...prev, general: 'Failed to load categories' })));
  }, []);

  const setField = useCallback(<K extends keyof RuleFormData>(field: K, value: RuleFormData[K]) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    // Clear field-specific error when user starts typing
    if (field === 'name' || field === 'content') {
      setErrors(prev => ({ ...prev, [field]: undefined }));
    }
  }, []);

  const setTriggers = useCallback((triggers: Trigger[]) => {
    setFormData(prev => ({ ...prev, triggers }));
  }, []);

  const resetErrors = useCallback(() => {
    setErrors({});
  }, []);

  const validate = useCallback((): boolean => {
    const newErrors: RuleFormErrors = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Name is required';
    } else if (formData.name.length > 200) {
      newErrors.name = 'Name must be under 200 characters';
    }

    if (!formData.content.trim()) {
      newErrors.content = 'Content is required';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  }, [formData.name, formData.content]);

  const handleSubmit = useCallback(async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validate()) {
      return;
    }

    try {
      setSaving(true);
      setErrors({});

      const request: CreateRuleRequest = {
        name: formData.name.trim(),
        content: formData.content.trim(),
        description: formData.description.trim() || undefined,
        target_layer: formData.targetLayer,
        category_id: formData.categoryId || undefined,
        priority_weight: formData.priorityWeight,
        overridable: formData.overridable,
        effective_start: formData.effectiveStart || undefined,
        effective_end: formData.effectiveEnd || undefined,
        tags: formData.tags ? formData.tags.split(',').map((t) => t.trim()).filter(Boolean) : undefined,
        team_id: teamId,
        triggers: formData.triggers.map((t) => ({
          type: t.type,
          pattern: t.pattern,
          context_types: t.contextTypes,
          tags: t.tags,
        })),
        enforcement_mode: formData.enforcementMode,
        temporary_timeout_hours: formData.temporaryTimeoutHours,
      };

      if (rule?.id) {
        await updateRule(rule.id, request);
      } else {
        await createRule(request);
      }

      onSuccess();
    } catch (err) {
      setErrors({
        general: err instanceof Error ? err.message : 'Failed to save rule',
      });
    } finally {
      setSaving(false);
    }
  }, [formData, teamId, rule?.id, validate, onSuccess]);

  return {
    formData,
    errors,
    categories,
    saving,
    setField,
    setTriggers,
    validate,
    handleSubmit,
    resetErrors,
  };
}
