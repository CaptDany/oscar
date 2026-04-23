import { useEffect, useRef } from 'preact/hooks';

interface ModalProps {
  id: string;
  title: string;
  size?: 'sm' | 'md' | 'lg' | 'xl';
}

const sizeClasses = {
  sm: 'max-w-md',
  md: 'max-w-lg',
  lg: 'max-w-2xl',
  xl: 'max-w-4xl',
};

interface PropsWithChildren {
  children?: any;
}

export function Modal({ id, title, size = 'md', children }: ModalProps & PropsWithChildren) {
  const overlayRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        const modal = document.getElementById(id);
        if (modal) {
          modal.classList.add('hidden');
          const backdrop = modal.previousElementSibling;
          if (backdrop) backdrop.remove();
        }
      }
    };

    const modal = document.getElementById(id);
    if (modal && !modal.classList.contains('hidden')) {
      document.addEventListener('keydown', handleEscape);
      document.body.style.overflow = 'hidden';
    }

    return () => {
      document.removeEventListener('keydown', handleEscape);
      document.body.style.overflow = 'unset';
    };
  }, [id]);

  const handleOverlayClick = (e: MouseEvent) => {
    if (e.target === overlayRef.current) {
      const modal = document.getElementById(id);
      if (modal) modal.classList.add('hidden');
      overlayRef.current?.parentElement?.previousElementSibling?.remove();
    }
  };

  return (
    <>
      <div
        ref={overlayRef}
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
        onClick={handleOverlayClick}
      />
      <div
        id={id}
        class="fixed inset-0 z-[51] hidden items-center justify-center p-4"
      >
        <div
          class={`bg-surface-container rounded-xl shadow-2xl w-full ${sizeClasses[size]} max-h-[90vh] flex flex-col overflow-hidden`}
        >
          <div class="flex items-center justify-between p-6 border-b border-outline/10">
            <h2 class="text-lg font-semibold text-on-surface">{title}</h2>
            <button
              onClick={() => {
                const modal = document.getElementById(id);
                if (modal) modal.classList.add('hidden');
              }}
              class="text-on-surface-variant hover:text-on-surface hover:bg-surface-container-low transition-colors p-2 rounded-lg"
            >
              <span class="material-symbols-outlined">close</span>
            </button>
          </div>
          <div class="p-6 overflow-y-auto flex-1">
            {children}
          </div>
        </div>
      </div>
    </>
  );
}

export function openModal(id: string) {
  const modal = document.getElementById(id);
  if (modal) {
    modal.classList.remove('hidden');
    modal.classList.add('flex');
  }
}

export function closeModal(id: string) {
  const modal = document.getElementById(id);
  if (modal) {
    modal.classList.add('hidden');
    modal.classList.remove('flex');
  }
}