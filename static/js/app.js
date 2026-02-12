function initDarkMode() {
    const isDark = localStorage.getItem('darkMode') === 'true' ||
        (!localStorage.getItem('darkMode') && window.matchMedia('(prefers-color-scheme: dark)').matches);
    document.documentElement.classList.toggle('dark', isDark);
}

function toggleDarkMode() {
    const isDark = document.documentElement.classList.toggle('dark');
    localStorage.setItem('darkMode', isDark);
}

function switchTab(status) {
    document.querySelectorAll('[data-column]').forEach(col => {
        col.classList.add('hidden');
        col.classList.remove('flex');
    });
    document.querySelector(`[data-column="${status}"]`).classList.remove('hidden');
    document.querySelector(`[data-column="${status}"]`).classList.add('flex');

    document.querySelectorAll('[data-tab]').forEach(tab => {
        tab.classList.remove('bg-primary-600', 'text-white');
        tab.classList.add('bg-gray-200', 'dark:bg-gray-700', 'text-gray-700', 'dark:text-gray-200');
    });
    const activeTab = document.querySelector(`[data-tab="${status}"]`);
    activeTab.classList.add('bg-primary-600', 'text-white');
    activeTab.classList.remove('bg-gray-200', 'dark:bg-gray-700', 'text-gray-700', 'dark:text-gray-200');
}

function initSortable() {
    document.querySelectorAll('.sortable-column').forEach(column => {
        if (column._sortable) return;
        column._sortable = Sortable.create(column, {
            group: 'kanban',
            animation: 150,
            ghostClass: 'sortable-ghost',
            dragClass: 'sortable-drag',
            handle: '.task-card',
            draggable: '.task-card',
            onEnd: function(evt) {
                const taskId = evt.item.dataset.taskId;
                const newStatus = evt.to.dataset.status;
                const children = Array.from(evt.to.children).filter(el => el.classList.contains('task-card'));
                const newIndex = children.indexOf(evt.item);

                let beforeId = 0;
                let afterId = 0;

                if (newIndex > 0) {
                    afterId = parseInt(children[newIndex - 1].dataset.taskId) || 0;
                }
                if (newIndex < children.length - 1) {
                    beforeId = parseInt(children[newIndex + 1].dataset.taskId) || 0;
                }

                fetch(`/tasks/${taskId}/move`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        status: newStatus,
                        after_id: afterId,
                        before_id: beforeId
                    })
                }).then(res => {
                    if (!res.ok) {
                        showToast('Failed to move task', 'error');
                        location.reload();
                    }
                }).catch(() => {
                    showToast('Failed to move task', 'error');
                    location.reload();
                });

                updateColumnCounts();
            }
        });
    });
}

function updateColumnCounts() {
    document.querySelectorAll('.sortable-column').forEach(column => {
        const count = column.querySelectorAll('.task-card').length;
        const countEl = column.parentElement.querySelector('.task-count');
        if (countEl) countEl.textContent = count;
    });
}

function showToast(message, type) {
    const toast = document.createElement('div');
    toast.className = `toast toast-${type || 'success'}`;
    toast.textContent = message;
    document.body.appendChild(toast);
    setTimeout(() => {
        toast.style.animation = 'slide-out 0.3s ease-in forwards';
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

function openModal(modalId) {
    document.getElementById(modalId).classList.remove('hidden');
}

function closeModal(modalId) {
    document.getElementById(modalId).classList.add('hidden');
}

document.addEventListener('DOMContentLoaded', () => {
    initDarkMode();
    initSortable();
});

document.addEventListener('htmx:afterSwap', () => {
    initSortable();
    updateColumnCounts();
});

document.addEventListener('htmx:afterSettle', () => {
    initSortable();
    updateColumnCounts();
});

document.addEventListener('htmx:afterRequest', (evt) => {
    const triggerHeader = evt.detail.xhr?.getResponseHeader('HX-Trigger');
    if (triggerHeader) {
        try {
            const triggers = JSON.parse(triggerHeader);
            if (triggers.showToast) {
                showToast(triggers.showToast.message, triggers.showToast.type);
            }
        } catch(e) {
        }
    }

    var mc = document.getElementById('modal-container');
    if (mc) mc.innerHTML = '';
    var cm = document.getElementById('create-modal');
    if (cm) cm.classList.add('hidden');
});
