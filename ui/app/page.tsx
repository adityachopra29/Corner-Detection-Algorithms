"use client";

import { fastJpeg, harris, shiTomashi, uploadImage } from "@/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Upload } from "lucide-react";
import { useState } from "react";
import Image from "next/image";

export default function Home() {
  const [image, setImage] = useState<File | null>(null);
  const [imageUploaded, setImageUploaded] = useState(false);

  const [fastLoading, setFastLoading] = useState(false);
  const [harrisLoading, setHarrisLoading] = useState(false);
  const [shiTomashiLoading, setShiTomashiLoading] = useState(false);

  const [outputURL, setOutputURL] = useState<string | null>(null);

  return (
    <div className="container mx-auto py-6 space-y-4">
      <h1 className="text-4xl tracking-tight">Corner Detection Algorithms</h1>
      <div className="flex flex-row gap-4 items-end">
        <div className="flex-1 self-end">
          <Label htmlFor="file">Upload Image</Label>
          <Input
            id="file"
            type="file"
            accept=".jpg"
            onChange={(e) => {
              setImage(e.target.files ? e.target.files[0] : null);
              setImageUploaded(false);
              setOutputURL(null);
            }}
          />
        </div>
        <Button
          size={"icon"}
          disabled={!image}
          onClick={async () => {
            if (image) {
              const response = await uploadImage(image);
              if (response.status === 200) {
                console.log("Image uploaded successfully");
                setImageUploaded(true);
              }
            }
          }}
        >
          <Upload />
        </Button>
      </div>
      {imageUploaded && (
        <p className="mx-auto text-emerald-500 font-medium tracking-tight text-sm">
          Image Uploaded Successfully :)
        </p>
      )}
      <div className="flex gap-4">
        <Button
          disabled={!imageUploaded || fastLoading}
          onClick={async () => {
            setFastLoading(true);
            setOutputURL(null);
            const path = await fastJpeg();
            if (path) setOutputURL(path);
            setFastLoading(false);
          }}
        >
          {fastLoading ? "Processing..." : "FAST Corner Detection"}
        </Button>
        <Button
          disabled={!imageUploaded || harrisLoading}
          onClick={async () => {
            setHarrisLoading(true);
            setOutputURL(null);
            const path = await harris();
            if (path) setOutputURL(path);
            setHarrisLoading(false);
          }}
        >
          {harrisLoading ? "Processing..." : "Harris Corner Detection"}
        </Button>
        <Button
          disabled={!imageUploaded || shiTomashiLoading}
          onClick={async () => {
            setShiTomashiLoading(true);
            setOutputURL(null);
            const path = await shiTomashi();
            if (path) setOutputURL(path);
            setShiTomashiLoading(false);
          }}
        >
          {shiTomashiLoading ? "Processing..." : "Shi Tomashi Corner Detection"}
        </Button>
      </div>

      <div>
        {outputURL && (
          <Image src={outputURL} height={800} width={1200} alt="output" />
        )}
      </div>
    </div>
  );
}
