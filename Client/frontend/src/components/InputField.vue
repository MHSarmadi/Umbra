<script setup lang="ts">
import { computed, ref } from 'vue'

interface Props {
	modelValue: string | number | null | undefined
	type?: string
	label?: string
	placeholder?: string
	helperText?: string
	errorText?: string
	size?: 'small' | 'medium' | 'large'
	id?: string
	name?: string
	autocomplete?: string
	inputmode?: 'none' | 'text' | 'tel' | 'url' | 'email' | 'numeric' | 'decimal' | 'search'
	disabled?: boolean
	readonly?: boolean
	required?: boolean
	clearable?: boolean
	checkoutable?: boolean
	min?: string | number
	max?: string | number
	step?: string | number
	minlength?: number
	maxlength?: number
	pattern?: string
	spellcheck?: boolean
	modelModifiers?: Record<string, boolean>
}

const props = withDefaults(defineProps<Props>(), {
	type: 'text',
	size: 'medium',
	autocomplete: 'off',
	clearable: false,
	spellcheck: false,
	placeholder: ''
})

const emit = defineEmits<{
	(e: 'update:modelValue', value: string | number | null): void
	(e: 'focus', event: FocusEvent): void
	(e: 'blur', event: FocusEvent): void
	(e: 'enter', value: string | number | null): void
	(e: 'checkout', value: string | number): void
}>()

const isFocused = ref(false)
const showPassword = ref(false)

const generatedId = `input-field-${Math.random().toString(36).slice(2, 10)}`
const inputId = computed(() => props.id || generatedId)
const hasValue = computed(() => String(props.modelValue ?? '').length > 0)
const isPasswordType = computed(() => props.type === 'password')
const inputType = computed(() => {
	if (isPasswordType.value) {
		return showPassword.value ? 'text' : 'password'
	}
	return props.type
})
const hasError = computed(() => Boolean(props.errorText))

function coerceValue(raw: string): string | number | null {
	if (props.modelModifiers?.number || props.type === 'number') {
		if (raw.trim() === '') return null
		const num = Number(raw)
		return Number.isFinite(num) ? num : null
	}
	return raw
}

function sanitizeByInputMode(raw: string): string {
	if (props.inputmode === 'numeric') {
		return raw.replace(/\D+/g, '')
	}

	if (props.inputmode === 'decimal') {
		const normalized = raw.replace(/,/g, '.').replace(/[^0-9.]/g, '')
		const [whole = '', ...fractionParts] = normalized.split('.')
		return fractionParts.length ? `${whole}.${fractionParts.join('')}` : whole
	}

	return raw
}

function onInput(event: Event) {
	const target = event.target as HTMLInputElement
	const sanitized = sanitizeByInputMode(target.value)
	if (sanitized !== target.value) {
		target.value = sanitized
	}
	emit('update:modelValue', coerceValue(sanitized))
}

function onFocus(event: FocusEvent) {
	isFocused.value = true
	emit('focus', event)
}

function onBlur(event: FocusEvent) {
	isFocused.value = false
	emit('blur', event)
}

function clearValue() {
	emit('update:modelValue', props.type === 'number' ? null : '')
}

function checkout() {
	if (props.checkoutable && hasValue.value) {
		emit('checkout', props.modelValue!)
	}
}

function togglePasswordVisibility() {
	showPassword.value = !showPassword.value
}

function onEnter() {
	emit('enter', props.modelValue ?? null)
}
</script>

<template>
	<label
		class="input-field"
		:class="[
			`size-${size}`,
			{
				'is-focused': isFocused,
				'has-value': hasValue,
				'has-error': hasError,
				'is-disabled': disabled,
				'has-prefix': !!$slots.prefix,
				'has-label': !!label
			}
		]"
		:for="inputId"
	>
		<div class="field-shell">
			<span v-if="$slots.prefix" class="prefix">
				<slot name="prefix" />
			</span>

			<div class="input-stack">
				<input
					:id="inputId"
					:name="name"
					class="native-input"
					:type="inputType"
					:value="modelValue ?? ''"
					:placeholder="placeholder"
					:autocomplete="autocomplete"
					:inputmode="inputmode"
					:disabled="disabled"
					:readonly="readonly"
					:required="required"
					:min="min"
					:max="max"
					:step="step"
					:minlength="minlength"
					:maxlength="maxlength"
					:pattern="pattern"
					:spellcheck="spellcheck"
					@input="onInput"
					@focus="onFocus"
					@blur="onBlur"
					@keyup.enter="onEnter"
				/>

				<span v-if="label" class="floating-label">
					{{ label }}
					<span v-if="required" class="required-mark">*</span>
				</span>
			</div>

			<div v-if="$slots.suffix || clearable || isPasswordType || checkoutable" class="right-controls">
				<span v-if="$slots.suffix" class="suffix">
					<slot name="suffix" />
				</span>

				<button
					v-if="clearable && hasValue && !disabled && !readonly"
					type="button"
					class="utility-btn"
					@click="clearValue"
					aria-label="Clear input"
				>
					Clear
				</button>

				<button
					v-if="checkoutable && hasValue && !disabled && !readonly"
					type="button"
					class="utility-btn colorful"
					@click="checkout"
					aria-label="Checkout CAPTCHA"
				>
					Checkout
				</button>

				<button
					v-if="isPasswordType && !disabled"
					type="button"
					class="utility-btn"
					@click="togglePasswordVisibility"
					:aria-label="showPassword ? 'Hide password' : 'Show password'"
				>
					{{ showPassword ? 'Hide' : 'Show' }}
				</button>
			</div>
		</div>

		<p v-if="hasError || helperText" class="message" :class="{ error: hasError }">
			{{ errorText || helperText }}
		</p>
	</label>
</template>

<style scoped lang="less">
@import "../style.less";

.input-field {
	display: flex;
	flex-direction: column;
	gap: 0.4rem;
	width: 100%;
	box-sizing: border-box;

	.field-shell {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		background: linear-gradient(180deg, #ffffff05 0%, #0000001f 100%), var(--dark-secondary-bg);
		border: 1px solid var(--hover-bg);
		border-radius: var(--border-radius-lg);
		padding: 0 0.85rem;
		box-shadow: inset 0 1px 0 #ffffff0d;
		transition: border-color 0.22s ease, box-shadow 0.22s ease, transform 0.22s ease;
	}

	.input-stack {
		position: relative;
		flex: 1;
		min-width: 0;
	}

	.native-input {
		width: 100%;
		border: 0;
		background: transparent;
		color: var(--text-color);
		font: inherit;
		padding: 1.18rem 0 0.6rem;
		line-height: 1.2;

		&::placeholder {
			color: color-mix(in srgb, var(--comment-color) 78%, #ffffff 22%);
		}
	}

	&:not(.has-label) .native-input {
		padding: 0.85rem 0;
	}

	.floating-label {
		position: absolute;
		left: 0;
		top: 50%;
		transform: translateY(-50%);
		font-size: 0.95rem;
		color: var(--comment-color);
		// text-transform: uppercase;
		pointer-events: none;
		transition: top 0.2s ease-out, transform 0.2s ease-out, font-size 0.2s ease-out, color 0.2s ease;
	}

	.required-mark {
		color: var(--main-highlight-color-4);
	}

	&.is-focused .floating-label,
	&.has-value .floating-label {
		top: 0.3rem;
		transform: translateY(0%);
		font-size: 0.6rem;
		letter-spacing: 0.05em;
		color: var(--main-highlight-color-4);
	}

	.prefix,
	.suffix {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		color: var(--comment-color);
		flex-shrink: 0;
	}

	.right-controls {
		display: inline-flex;
		align-items: center;
		gap: 0.4rem;
		flex-shrink: 0;
	}

	.utility-btn {
		border: 0;
		background: var(--hover-bg);
		color: var(--text-color);
		border-radius: var(--border-radius-sm);
		padding: 0.28rem 0.48rem;
		font: inherit;
		font-size: 0.78rem;
		cursor: pointer;
		transition: background-color 0.2s ease, color 0.2s ease;

		&:hover {
			background: var(--bright-bg);
		}

		&:active {
			background: var(--main-highlight-color);
		}

		&.colorful {
			background: var(--main-highlight-color);
			color: var(--text-color);
			&:hover {
				background: var(--main-highlight-color-2);
			}
			&:active {
				background: var(--secondary-highlight-color);
			}
		}
	}

	&.is-focused .field-shell {
		border-color: var(--main-highlight-color-2);
		box-shadow: 0 0 0 3px #00897b2b, inset 0 1px 0 #ffffff14;
		transform: translateY(-1px);
	}

	&.has-error {
		.floating-label {
			color: #f97373;
		}

		.field-shell {
			border-color: #db5c5c;
			box-shadow: 0 0 0 3px #db5c5c2c;
		}
	}

	&.is-disabled {
		opacity: 0.72;
		pointer-events: none;
	}

	.message {
		margin: 0;
		font-size: 0.84rem;
		color: var(--comment-color);
		padding-left: 0.2rem;
	}

	.message.error {
		color: #f08a8a;
	}

	&.size-small {
		.floating-label {
			font-size: 0.88rem;
		}

		.native-input {
			padding: 0.95rem 0 0.28rem;
			font-size: 0.94rem;
		}
	}

	&.size-large {
		.floating-label {
			font-size: 1rem;
		}

		.native-input {
			padding: 1.2rem 0 0.44rem;
			font-size: 1.08rem;
		}

		.utility-btn {
			font-size: 0.82rem;
		}
	}
}
</style>
