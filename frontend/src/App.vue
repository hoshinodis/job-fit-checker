<script setup lang="ts">
import { ref, reactive } from 'vue'
import { postMatch, getMatchStatus } from '@/api/client'
import type { MatchResultOutput } from '@/types'

const profile = reactive({
  preferred_languages: '',
  avoid_languages: '',
  interests: '',
  low_interests: '',
  work_style: '',
  desired_compensation: '',
  notes: '',
})

const jobInputType = ref<'text' | 'url'>('text')
const jobInputValue = ref('')

const isRunning = ref(false)
const jobStatus = ref('')
const result = ref<MatchResultOutput | null>(null)
const errorMsg = ref('')
const copied = ref(false)

function splitComma(s: string): string[] {
  return s.split(/[,、]/).map(v => v.trim()).filter(Boolean)
}

async function submit() {
  errorMsg.value = ''
  result.value = null
  copied.value = false

  if (!jobInputValue.value.trim()) {
    errorMsg.value = '求人情報を入力してください'
    return
  }

  isRunning.value = true
  jobStatus.value = 'queued'

  try {
    const res = await postMatch({
      profile: {
        preferred_languages: splitComma(profile.preferred_languages),
        avoid_languages: splitComma(profile.avoid_languages),
        interests: splitComma(profile.interests),
        low_interests: splitComma(profile.low_interests),
        work_style: splitComma(profile.work_style),
        desired_compensation: profile.desired_compensation,
        notes: profile.notes,
      },
      job_input: {
        type: jobInputType.value,
        value: jobInputValue.value,
      },
    })

    const jobId = res.id
    await pollResult(jobId)
  } catch (e: any) {
    errorMsg.value = e.message || '送信に失敗しました'
  } finally {
    isRunning.value = false
  }
}

async function pollResult(id: string) {
  const interval = 2000
  const maxAttempts = 120
  for (let i = 0; i < maxAttempts; i++) {
    await new Promise(r => setTimeout(r, interval))
    const status = await getMatchStatus(id)
    jobStatus.value = status.status

    if (status.status === 'done' && status.result) {
      result.value = status.result
      return
    }
    if (status.status === 'failed') {
      const msg = status.error || '判定に失敗しました'
      if (jobInputType.value === 'url') {
        errorMsg.value = msg + '\n\nURLからの取得に失敗した可能性があります。求人本文をテキストで貼り付けてお試しください。'
      } else {
        errorMsg.value = msg
      }
      return
    }
  }
  errorMsg.value = 'タイムアウトしました'
}

async function copyResult() {
  if (!result.value) return
  try {
    await navigator.clipboard.writeText(result.value.clipboard_text)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch {
    errorMsg.value = 'コピーに失敗しました'
  }
}
</script>

<template>
  <div class="container">
    <h1>求人マッチ度チェッカー</h1>

    <section class="section">
      <h2>プロフィール</h2>
      <div class="field">
        <label>得意言語（カンマ区切り）</label>
        <input v-model="profile.preferred_languages" placeholder="Go, Python" />
      </div>
      <div class="field">
        <label>避けたい言語（カンマ区切り）</label>
        <input v-model="profile.avoid_languages" placeholder="Java, Kotlin" />
      </div>
      <div class="field">
        <label>興味のある領域（カンマ区切り）</label>
        <input v-model="profile.interests" placeholder="推薦, マッチング" />
      </div>
      <div class="field">
        <label>興味の薄い領域（カンマ区切り）</label>
        <input v-model="profile.low_interests" placeholder="IaC中心の業務" />
      </div>
      <div class="field">
        <label>希望する働き方（カンマ区切り）</label>
        <input v-model="profile.work_style" placeholder="柔軟な働き方, 裁量" />
      </div>
      <div class="field">
        <label>希望年収（任意）</label>
        <input v-model="profile.desired_compensation" placeholder="800万円以上" />
      </div>
      <div class="field">
        <label>自由記述</label>
        <textarea v-model="profile.notes" rows="3" placeholder="バックエンド中心だがプロダクト寄り。"></textarea>
      </div>
    </section>

    <section class="section">
      <h2>求人情報</h2>
      <div class="radio-group">
        <label><input type="radio" value="text" v-model="jobInputType" /> テキスト</label>
        <label><input type="radio" value="url" v-model="jobInputType" /> URL</label>
      </div>
      <div class="field" v-if="jobInputType === 'text'">
        <textarea v-model="jobInputValue" rows="8" placeholder="求人票の本文を貼り付けてください"></textarea>
      </div>
      <div class="field" v-else>
        <input v-model="jobInputValue" type="url" placeholder="https://example.com/job/123" />
      </div>
    </section>

    <button class="btn-primary" :disabled="isRunning" @click="submit">
      {{ isRunning ? '判定中...' : 'マッチ度を判定する' }}
    </button>

    <div v-if="isRunning" class="status">
      ステータス: {{ jobStatus }}
    </div>

    <div v-if="errorMsg" class="error">
      <pre>{{ errorMsg }}</pre>
    </div>

    <section v-if="result" class="result">
      <h2>判定結果</h2>
      <div class="score">マッチ度: <strong>{{ result.score }}</strong> / 100</div>
      <div class="summary"><strong>要約:</strong> {{ result.summary }}</div>

      <div v-if="result.pros.length">
        <h3>一致している点</h3>
        <ul><li v-for="p in result.pros" :key="p">{{ p }}</li></ul>
      </div>

      <div v-if="result.cons.length">
        <h3>懸念点</h3>
        <ul><li v-for="c in result.cons" :key="c">{{ c }}</li></ul>
      </div>

      <div v-if="result.questions_to_ask.length">
        <h3>面接で確認すべき質問</h3>
        <ul><li v-for="q in result.questions_to_ask" :key="q">{{ q }}</li></ul>
      </div>

      <div v-if="result.model" class="model">使用モデル: {{ result.model }}</div>

      <button class="btn-copy" @click="copyResult">
        {{ copied ? 'コピーしました！' : '結果をコピー' }}
      </button>
    </section>
  </div>
</template>

<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; color: #333; }
.container { max-width: 720px; margin: 0 auto; padding: 2rem 1rem; }
h1 { text-align: center; margin-bottom: 1.5rem; color: #1a1a2e; }
h2 { margin-bottom: 0.75rem; color: #16213e; border-bottom: 2px solid #0f3460; padding-bottom: 0.25rem; }
h3 { margin: 0.75rem 0 0.25rem; color: #0f3460; }
.section { background: #fff; border-radius: 8px; padding: 1.25rem; margin-bottom: 1.25rem; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
.field { margin-bottom: 0.75rem; }
.field label { display: block; font-size: 0.875rem; font-weight: 600; margin-bottom: 0.25rem; }
.field input, .field textarea { width: 100%; padding: 0.5rem; border: 1px solid #ccc; border-radius: 4px; font-size: 0.95rem; }
.field textarea { resize: vertical; }
.radio-group { margin-bottom: 0.75rem; display: flex; gap: 1.5rem; }
.radio-group label { font-weight: 500; cursor: pointer; }
.btn-primary { display: block; width: 100%; padding: 0.75rem; background: #0f3460; color: #fff; border: none; border-radius: 6px; font-size: 1rem; font-weight: 600; cursor: pointer; margin-bottom: 1rem; }
.btn-primary:disabled { background: #999; cursor: not-allowed; }
.btn-primary:hover:not(:disabled) { background: #16213e; }
.btn-copy { margin-top: 1rem; padding: 0.5rem 1.25rem; background: #e94560; color: #fff; border: none; border-radius: 4px; cursor: pointer; font-weight: 600; }
.btn-copy:hover { background: #c73e54; }
.status { text-align: center; padding: 0.5rem; color: #0f3460; font-weight: 500; }
.error { background: #fff0f0; border: 1px solid #e94560; border-radius: 6px; padding: 1rem; margin-bottom: 1rem; }
.error pre { white-space: pre-wrap; color: #c0392b; font-size: 0.9rem; }
.result { background: #fff; border-radius: 8px; padding: 1.25rem; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
.score { font-size: 1.5rem; margin-bottom: 0.5rem; }
.score strong { color: #0f3460; }
.summary { margin-bottom: 0.75rem; line-height: 1.5; }
.model { margin-top: 0.75rem; font-size: 0.85rem; color: #888; }
ul { padding-left: 1.25rem; margin-bottom: 0.5rem; }
li { margin-bottom: 0.25rem; line-height: 1.4; }
</style>
