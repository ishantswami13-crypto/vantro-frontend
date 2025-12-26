"use client";

import { motion, AnimatePresence } from "framer-motion";
import { cn } from "@/lib/utils";

export function Toast({
  show,
  children,
  className,
}: {
  show: boolean;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <AnimatePresence>
      {show && (
        <motion.div
          initial={{ opacity: 0, y: -10, scale: 0.98 }}
          animate={{ opacity: 1, y: 0, scale: 1 }}
          exit={{ opacity: 0, y: -10, scale: 0.98 }}
          transition={{ duration: 0.18 }}
          className={cn(
            "fixed top-5 left-1/2 z-50 w-[92%] max-w-lg -translate-x-1/2 rounded-2xl border border-white/10 bg-black/60 backdrop-blur-xl px-4 py-3 shadow-[0_20px_60px_-30px_rgba(0,0,0,0.9)]",
            className
          )}
        >
          {children}
        </motion.div>
      )}
    </AnimatePresence>
  );
}
