# Инструкция по обновлению документации VEDO Core

## Когда обновлять документацию

- Изменение функциональности, API, CLI, deployment или user workflow
- Добавление или изменение артефактов в `human/artifacts/`
- Изменение ADR, C4-диаграмм, архитектурных решений
- Pull request, который меняет поведение системы, **должен** обновлять документацию в том же PR

## Процесс обновления

### Шаг 1: Инвентаризация изменений

Сравните текущие артефакты с содержимым страниц документации:

```bash
# Проверка изменений в артефактах
git diff HEAD -- human/artifacts/

# Проверка изменений в коде
git diff HEAD -- services/ frontend/ cli/
```

### Шаг 2: Определение затрагиваемых компонентов

| Изменение | Компонент |
|-----------|-----------|
| API, протоколы, архитектура | developer-guide, integrator-guide |
| CLI, backup, deployment, мониторинг | admin-guide |
| Интерфейс, сценарии пользователя | user-guide |
| ADR, C4-диаграммы | developer-guide |

### Шаг 3: Обновление AsciiDoc-страниц

1. Откройте соответствующую страницу в `docs/antora/{component}/modules/ROOT/pages/`
2. Обновите содержимое в соответствии с изменениями
3. Проверьте соответствие правилам качества (см. `review-checklist.md`)

### Шаг 4: Проверка сборки

```bash
# Установка Antora (если не установлен)
npm install -g @antora/cli @antora/site-generator-default

# Сборка всех компонентов
cd docs/antora
antora generate playbook-user.yml
antora generate playbook-dev.yml
antora generate playbook-admin.yml
antora generate playbook-integrator.yml
```

### Шаг 5: Проверка целостности

- Все cross-reference ссылки работают
- Навигация (nav.adoc) полная и логичная
- Нет битых ссылок
- Команды в примерах проверены на выполнимость

## Команды CI

В GitLab CI документация собирается на стадии `docs`:

```yaml
docs:
  stage: docs
  script:
    - antora generate docs/antora/playbook-user.yml
    - antora generate docs/antora/playbook-dev.yml
    - antora generate docs/antora/playbook-admin.yml
    - antora generate docs/antora/playbook-integrator.yml
  artifacts:
    paths:
      - build/site/
```

## Проверки в CI

| Проверка | Инструмент | Команда |
|----------|-----------|---------|
| Spell/style lint | typos, vale | `typos docs/antora/` |
| Link checker | htmltest | `htmltest build/site/` |
| Antora build | antora | `antora generate playbook-*.yml` |
| Командные snippets | shell smoke | `bash -n docs/**/*.adoc` |

## Исключения

Отдельное обновление документации (не в том же PR, что изменение кода) допускается только как emergency fix не позднее 1 рабочего дня после изменения кода.
