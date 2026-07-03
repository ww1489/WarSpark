import { useCallback, useRef, useState } from "react";
import { useNavigate } from "@tanstack/react-router";
import { createImageSearchJob } from "~/lib/api";
import { Card } from "~/components/ui/Card";
import { Button } from "~/components/ui/Button";
import { ErrorCard } from "~/components/ui/ErrorCard";

const TH_OPTIONS = Array.from({ length: 15 }, (_, i) => i + 4);

export default function FindLayoutPage() {
  const [file, setFile] = useState<File | null>(null);
  const [thLevel, setThLevel] = useState<number | "">("");
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const fileRef = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();

  const handleFile = useCallback((f: File | null) => {
    setFile(f);
    setError(null);
  }, []);

  const handleUpload = async () => {
    if (!file) return;
    setUploading(true);
    setError(null);
    try {
      const job = await createImageSearchJob(file, thLevel ? Number(thLevel) : undefined);
      navigate({ to: "/find-layout/results/$jobId", params: { jobId: job.job_id } });
    } catch (e) {
      setError(e instanceof Error ? e.message : "上传失败");
    } finally {
      setUploading(false);
    }
  };

  return (
    <div className="max-w-xl mx-auto py-14 px-4 space-y-8">
      <div className="text-center">
        <h1 className="text-3xl font-bold tracking-tight text-text-primary">找阵型</h1>
        <p className="text-sm text-text-secondary mt-2">上传游戏截图，AI 自动识别并匹配相似阵型</p>
      </div>

      <Card className="shadow-card">
        <Card.Body className="p-8 space-y-8">
          {/* Upload Zone */}
          <div
            onDragOver={(e) => e.preventDefault()}
            onDrop={(e) => { e.preventDefault(); handleFile(e.dataTransfer.files?.[0] ?? null); }}
            onClick={() => fileRef.current?.click()}
            className={`relative border-2 border-dashed rounded-2xl p-12 text-center cursor-pointer transition-all duration-200 ${
              file
                ? "border-primary bg-primary-bg/10"
                : "border-border hover:border-primary/40 hover:bg-primary-bg/5"
            }`}
          >
            <input ref={fileRef} type="file" accept="image/*" className="hidden" onChange={(e) => handleFile(e.target.files?.[0] ?? null)} />
            {file ? (
              <div className="space-y-3">
                <div className="w-16 h-16 mx-auto rounded-2xl bg-primary-bg flex items-center justify-center text-2xl">🖼</div>
                <div>
                  <p className="text-sm font-semibold text-text-primary">{file.name}</p>
                  <p className="text-xs text-text-muted mt-1">{Math.round(file.size / 1024)} KB</p>
                </div>
                <Button variant="ghost" size="sm" onClick={(e) => { e.stopPropagation(); handleFile(null); }}>
                  重新选择
                </Button>
              </div>
            ) : (
              <div className="space-y-3">
                <div className="w-20 h-20 mx-auto rounded-2xl bg-primary-bg flex items-center justify-center text-3xl">📷</div>
                <p className="text-text-secondary text-sm font-medium">点击或拖拽截图到此处</p>
                <p className="text-text-muted text-xs">PNG · JPG · WEBP，不超过 10MB</p>
              </div>
            )}
          </div>

          {/* TH Selector */}
          <div>
            <p className="text-sm font-semibold text-text-primary mb-4">大本营等级 <span className="text-text-muted font-normal">（可选）</span></p>
            <div className="grid grid-cols-4 sm:grid-cols-8 gap-2">
              <button
                type="button"
                onClick={() => setThLevel("")}
                className={`py-2 px-1 text-xs font-medium rounded-xl border transition-all ${
                  thLevel === ""
                    ? "bg-primary text-white border-primary shadow-sm"
                    : "border-border text-text-secondary hover:border-primary/40 hover:text-text-primary"
                }`}
              >
                自动
              </button>
              {TH_OPTIONS.map((n) => (
                <button
                  key={n}
                  type="button"
                  onClick={() => setThLevel(n)}
                  className={`py-2 px-1 text-xs font-semibold rounded-xl border transition-all ${
                    thLevel === n
                      ? "bg-primary text-white border-primary shadow-sm"
                      : "border-border text-text-secondary hover:border-primary/40 hover:text-text-primary"
                  }`}
                >
                  TH{n}
                </button>
              ))}
            </div>
          </div>

          {error && <ErrorCard message={error} />}

          <Button onClick={handleUpload} loading={uploading} disabled={!file} className="w-full h-12 text-sm font-semibold rounded-xl" size="lg">
            {uploading ? "上传分析中..." : "开始搜索"}
          </Button>
        </Card.Body>
      </Card>
    </div>
  );
}
