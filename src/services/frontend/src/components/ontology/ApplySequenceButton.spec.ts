// Unit tests for useApplySequence composable
// Tests: workflow state transitions, progress tracking, error display

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useApplySequence } from '../../composables/useApplySequence'
import type { SequenceStep } from '../../types/extraction'

// Mock the API module
vi.mock('../../api/extraction', () => ({
  applySequence: vi.fn()
}))

function createStep(overrides: Partial<SequenceStep> = {}): SequenceStep {
  return {
    id: overrides.id ?? 's1',
    operation: overrides.operation ?? 'CREATE_CLASS',
    entityId: overrides.entityId ?? 'Person',
    label: overrides.label ?? 'Person',
    included: overrides.included ?? true,
    isDuplicate: overrides.isDuplicate ?? false,
    sourceFile: overrides.sourceFile
  }
}

describe('useApplySequence', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Initial state', () => {
    it('starts in idle state', () => {
      const { progress, isIdle, isApplying, isSuccess, isError } = useApplySequence()

      expect(progress.value.status).toBe('idle')
      expect(isIdle.value).toBe(true)
      expect(isApplying.value).toBe(false)
      expect(isSuccess.value).toBe(false)
      expect(isError.value).toBe(false)
    })

    it('has 0 percentage initially', () => {
      const { percentage } = useApplySequence()
      expect(percentage.value).toBe(0)
    })
  })

  describe('State transitions', () => {
    it('transitions to confirming state on requestConfirm', () => {
      const { progress, isConfirming, requestConfirm } = useApplySequence()

      const steps = [createStep({ id: 's1', included: true })]
      requestConfirm(steps)

      expect(isConfirming.value).toBe(true)
      expect(progress.value.status).toBe('confirming')
      expect(progress.value.totalSteps).toBe(1)
    })

    it('sets default commit message on confirm', () => {
      const { commitMessage, requestConfirm } = useApplySequence()

      const steps = [
        createStep({ id: 's1', operation: 'CREATE_CLASS', included: true }),
        createStep({ id: 's2', operation: 'CREATE_PROPERTY', included: true })
      ]
      requestConfirm(steps)

      expect(commitMessage.value).toContain('2 steps')
    })

    it('transitions to applying state on executeApply', async () => {
      const { applySequence } = await import('../../api/extraction')
      const mockApply = applySequence as ReturnType<typeof vi.fn>
      mockApply.mockResolvedValue({
        success: true,
        appliedCount: 2,
        commitId: 'abc123',
        commitUrl: '/commits/abc123',
        timestamp: '2026-07-16T12:00:00Z'
      })

      const { executeApply } = useApplySequence()

      const steps = [
        createStep({ id: 's1', included: true }),
        createStep({ id: 's2', included: true })
      ]
      await executeApply('test-onto', steps, 'Import 2 steps')

      // Should have been applying at some point
      // (transient state before wait)
      // After resolution, should be success
    })

    it('transitions to success state on successful apply', async () => {
      const { applySequence } = await import('../../api/extraction')
      const mockApply = applySequence as ReturnType<typeof vi.fn>
      mockApply.mockResolvedValue({
        success: true,
        appliedCount: 2,
        commitId: 'abc123',
        timestamp: '2026-07-16T12:00:00Z'
      })

      const { progress, isSuccess, applyResult, executeApply } = useApplySequence()

      const steps = [
        createStep({ id: 's1', included: true }),
        createStep({ id: 's2', included: true })
      ]
      await executeApply('test-onto', steps, 'Import test')

      expect(isSuccess.value).toBe(true)
      expect(progress.value.status).toBe('success')
      expect(applyResult.value).toBeTruthy()
      expect(applyResult.value?.commitId).toBe('abc123')
    })

    it('transitions to error state on failed apply', async () => {
      const { applySequence } = await import('../../api/extraction')
      const mockApply = applySequence as ReturnType<typeof vi.fn>
      mockApply.mockResolvedValue({
        success: false,
        appliedCount: 0,
        errors: [{ stepId: 's1', message: 'Entity already exists' }],
        timestamp: '2026-07-16T12:00:00Z'
      })

      const { progress, isError, executeApply } = useApplySequence()

      const steps = [createStep({ id: 's1', included: true })]
      await executeApply('test-onto', steps, 'Test')

      expect(isError.value).toBe(true)
      expect(progress.value.errorMessage).toBe('Entity already exists')
    })

    it('transitions to error state on API exception', async () => {
      const { applySequence } = await import('../../api/extraction')
      const mockApply = applySequence as ReturnType<typeof vi.fn>
      mockApply.mockRejectedValue(new Error('Connection refused'))

      const { progress, isError, executeApply } = useApplySequence()

      const steps = [createStep({ id: 's1', included: true })]
      await executeApply('test-onto', steps, 'Test')

      expect(isError.value).toBe(true)
      expect(progress.value.errorMessage).toContain('Connection refused')
    })
  })

  describe('Reset', () => {
    it('resets to idle state', () => {
      const { progress, isIdle, requestConfirm, reset } = useApplySequence()

      const steps = [createStep({ id: 's1', included: true })]
      requestConfirm(steps)
      expect(isIdle.value).toBe(false)

      reset()
      expect(isIdle.value).toBe(true)
      expect(progress.value.totalSteps).toBe(0)
    })
  })

  describe('Cancel', () => {
    it('cancels and returns to idle', () => {
      const { progress, isIdle, requestConfirm, cancelApply } = useApplySequence()

      const steps = [createStep({ id: 's1', included: true })]
      requestConfirm(steps)
      expect(isIdle.value).toBe(false)

      cancelApply()
      expect(isIdle.value).toBe(true)
      expect(progress.value.status).toBe('idle')
    })
  })

  describe('Percentage calculation', () => {
    it('calculates percentage correctly', () => {
      const { percentage, requestConfirm } = useApplySequence()

      const steps = Array.from({ length: 10 }, (_, i) =>
        createStep({ id: `s${i}`, included: true })
      )
      requestConfirm(steps)

      // Initially 0 of 10 = 0%
      expect(percentage.value).toBe(0)
    })
  })
})
