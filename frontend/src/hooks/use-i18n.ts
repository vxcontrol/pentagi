import { useTranslation } from 'react-i18next';

export const useI18n = () => {
  const { t, i18n } = useTranslation();

  const changeLanguage = (lng: string) => {
    i18n.changeLanguage(lng);
    // Store in localStorage for persistence
    localStorage.setItem('i18nextLng', lng);
  };

  const currentLanguage = i18n.language || 'en';

  return {
    t,
    changeLanguage,
    currentLanguage,
    i18n,
  };
};

// Convenience hook for common translations
export const useTranslations = () => {
  const { t } = useTranslation();
  return t;
};