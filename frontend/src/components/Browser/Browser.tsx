import { useState, useMemo, useEffect, useCallback } from "react";
import { Gallery } from "react-grid-gallery";
import Lightbox from "yet-another-react-lightbox";
import Zoom from "yet-another-react-lightbox/plugins/zoom";
import Thumbnails from "yet-another-react-lightbox/plugins/thumbnails";
import "yet-another-react-lightbox/styles.css";
import "yet-another-react-lightbox/plugins/thumbnails.css";
import { AnimatePresence } from "framer-motion";

import {
  headerStyles,
  galleryWrapperStyles,
  galleryContainerStyles,
  wrapperStyles,
  emptyStateStyles,
} from "./Browser.css";
import type { Screenshot } from "@/generated/graphql";

type BrowserProps = {
  screenshots: Screenshot[];
};

const baseURL = "/api/v1";

interface CustomImage {
  src: string;
  width: number;
  height: number;
  caption: string;
  thumbnailWidth?: number;
  thumbnailHeight?: number;
}

export const Browser = ({ screenshots }: BrowserProps) => {
  const [lightboxIndex, setLightboxIndex] = useState<number>(-1);
  const [images, setImages] = useState<CustomImage[]>([]);
  const isLightboxOpen = lightboxIndex >= 0;

  const sortedScreenshots = useMemo(() => {
    return [...screenshots].sort((a, b) => 
      parseInt(b.id) - parseInt(a.id)
    );
  }, [screenshots]);

  const imageUrls = useMemo(() => {
    return sortedScreenshots.map(screenshot => ({
      src: `${baseURL}/flows/${screenshot.flowId}/screenshots/${screenshot.id}/file`,
      caption: screenshot.url || ""
    }));
  }, [sortedScreenshots]);

  useEffect(() => {
    const loadImageDetails = async () => {
      const loadedImages = await Promise.all(
        imageUrls.map(async ({ src, caption }) => {
          const img = new Image();
          return new Promise<CustomImage>((resolve) => {
            img.onload = () => {
              resolve({
                src,
                width: img.naturalWidth,
                height: img.naturalHeight,
                caption,
                thumbnailWidth: Math.min(320, img.naturalWidth),
                thumbnailHeight: Math.min(240, img.naturalHeight)
              });
            };
            img.src = src;
          });
        })
      );
      setImages(loadedImages);
    };

    loadImageDetails();
  }, [imageUrls]);

  const handleClose = useCallback(() => {
    setLightboxIndex(-1);
    if (document.activeElement instanceof HTMLElement) {
      document.activeElement.blur();
    }
    setTimeout(() => {
      const galleryElements = document.querySelectorAll('[role="grid"]');
      galleryElements.forEach(element => {
        if (element instanceof HTMLElement) {
          element.style.outline = 'none';
          element.blur();
        }
      });
    }, 0);
  }, []);

  const handleClick = (index: number) => {
    setLightboxIndex(index);
  };

  if (!screenshots.length) {
    return (
      <div className={wrapperStyles}>
        <div className={headerStyles}>Browser</div>
        <div className={emptyStateStyles}>
          No screenshots available
        </div>
      </div>
    );
  }

  return (
    <div className={wrapperStyles}>
      <div className={headerStyles}>
        Browser ({screenshots.length} screenshots)
      </div>
      
      <div className={galleryWrapperStyles}>
        <div className={galleryContainerStyles}>
          <Gallery
            images={images}
            onClick={(index) => handleClick(index)}
            enableImageSelection={false}
            rowHeight={240}
            margin={4}
            tileViewportStyle={() => ({
              width: '100%',
              height: '100%',
              overflow: 'hidden'
            })}
            thumbnailStyle={() => ({
              width: '100%',
              height: '100%',
              objectFit: 'cover' as const,
              cursor: 'pointer'
            })}
          />
        </div>
      </div>

      <AnimatePresence>
        {isLightboxOpen && (
          <Lightbox
            open={isLightboxOpen}
            close={handleClose}
            slides={images.map(img => ({ 
              src: img.src, 
              description: img.caption,
              width: img.width,
              height: img.height
            }))}
            index={lightboxIndex}
            plugins={[Zoom, Thumbnails]}
            animation={{ fade: 300 }}
            zoom={{
              maxZoomPixelRatio: 5,
            }}
            carousel={{
              finite: true
            }}
            render={{
              buttonPrev: () => null,
              buttonNext: () => null
            }}
          />
        )}
      </AnimatePresence>
    </div>
  );
};

export default Browser;
