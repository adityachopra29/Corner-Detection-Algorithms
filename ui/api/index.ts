const backend = "http://localhost:8080";

export const uploadImage = async (file: File) => {
  const formData = new FormData();
  formData.append("image", file);

  const response = await fetch(`${backend}/upload`, {
    method: "POST",
    body: formData,
  });

  return response;
};

export const fastJpeg = async () => {
  const response = await fetch(`${backend}/fast-jpeg`);
  if (response.status === 200) {
    console.log("FAST corner detection successful");
    const res = await response.json();
    return `${backend}/${res.path}`;
  } else {
    return null;
  }
};

export const harris = async () => {
  const response = await fetch(`${backend}/harris`);
  if (response.status === 200) {
    console.log("Harris corner detection successful");
    const res = await response.json();
    return `${backend}/${res.path}`;
  } else {
    return null;
  }
};

export const shiTomashi = async () => {
  const response = await fetch(`${backend}/shi-tomashi`);
  if (response.status === 200) {
    console.log("Shi Tomashi corner detection successful");
    const res = await response.json();
    return `${backend}/${res.path}`;
  } else {
    return null;
  }
};
